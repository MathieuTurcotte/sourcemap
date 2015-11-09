// Copyright (c) 2013 Mathieu Turcotte
// Licensed under the MIT license.

package sourcemap

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

const maxEntryValue = 5    // Maximum number of values within a single entry.
const lineSeparator = ';'  // Line separator in the "mappings" field.
const entrySeparator = ',' // Entry separator withing a line in the "mappings" field.

const defaultNumLines = 256  // Default number of lines to allocate per source map.
const defaultNumEntries = 16 // Default number of entries to allocate per line.

// A struct into which the source map JSON representation is unmarshalled.
type jsonSourceMap struct {
	Version        int
	File           string
	SourceRoot     string
	Sources        []string
	SourcesContent []string
	Names          []interface{}
	Mappings       string
}

// Reads a source map.
func Read(reader io.Reader) (s SourceMap, err error) {
	var jsonMap jsonSourceMap
	var lines []Line

	dec := json.NewDecoder(reader)
	if err = dec.Decode(&jsonMap); err != nil {
		return
	}

	if jsonMap.Version != 3 {
		err = fmt.Errorf("unsupported version: %v", s.Version)
	}

	if lines, err = parseMappings(jsonMap.Mappings); err != nil {
		return
	}

	// Populate the source map structure.
	s.Version = jsonMap.Version
	s.File = jsonMap.File
	s.SourceRoot = jsonMap.SourceRoot
	s.Sources = jsonMap.Sources
	s.SourcesContent = jsonMap.SourcesContent
	s.Names = make([]string, len(jsonMap.Names))
	for i, v := range jsonMap.Names {
		switch v := v.(type) {
		case string:
			s.Names[i] = v
		case int:
			s.Names[i] = fmt.Sprintf("%d", v)
		}
	}
	s.Mappings = lines

	return
}

// A struct to hold the state while parsing the mappings.
type mappingsParseState struct {
	gencol int // 0-based index of the generated column.
	source int // 0-based index into the "sources" list.
	line   int // 0-based starting line in the original source.
	column int // 0-based starting column of the line in the source.
	name   int // 0-based index into the "names" list.
}

func parseMappings(data string) (lines []Line, err error) {
	lines = make([]Line, 0, defaultNumLines)
	reader := strings.NewReader(data)
	state := mappingsParseState{}
	line := make(Line, 0, defaultNumEntries)
	for reader.Len() > 0 {
		if consume(reader, lineSeparator) {
			lines = append(lines, line)
			line = make(Line, 0, defaultNumEntries)
			state.gencol = 0
		} else {
			i := 0
			var values [maxEntryValue]int
			for !entryCompleted(reader) {
				if val, derr := decodeVQL(reader); derr != nil {
					err = derr
					return
				} else {
					values[i] = val
					i++
				}
			}

			entry, eerr := newEntry(&state, values, i)
			if eerr != nil {
				err = eerr
				return
			}

			line = append(line, entry)
			consume(reader, entrySeparator)
		}
	}

	return
}

// Attempts to consume the wanted character from the reader. Returns true if
// character was read successfully, false otherwise. If the character could
// not be read, then no characters are consumed from the reader.
func consume(r *strings.Reader, wanted byte) bool {
	if r.Len() > 0 {
		b, _ := r.ReadByte()
		if b == wanted {
			return true
		}
		r.UnreadByte()
	}
	return false
}

// Looks ahead to see if the next character completes the entry. Does not read
// any characters from the reader.
func entryCompleted(r *strings.Reader) bool {
	if r.Len() > 0 {
		b, _ := r.ReadByte()
		r.UnreadByte()
		return b == lineSeparator || b == entrySeparator
	}
	return true
}

func newEntry(s *mappingsParseState, values [maxEntryValue]int,
	numValues int) (e Entry, err error) {
	e = Entry{-1, -1, -1, -1, -1}

	switch numValues {
	case 1:
		// Unmapped section of the generated file.
		e.GeneratedColumn = values[0] + s.gencol
		s.gencol = e.GeneratedColumn
	case 4:
		// Mapped section of the generated file.
		e.GeneratedColumn = values[0] + s.gencol
		e.SourceFileId = values[1] + s.source
		e.SourceLine = values[2] + s.line
		e.SourceColumn = values[3] + s.column
		s.gencol = e.GeneratedColumn
		s.source = e.SourceFileId
		s.line = e.SourceLine
		s.column = e.SourceColumn
	case 5:
		// Mapped section of the generated file with an associated name.
		e.GeneratedColumn = values[0] + s.gencol
		e.SourceFileId = values[1] + s.source
		e.SourceLine = values[2] + s.line
		e.SourceColumn = values[3] + s.column
		e.NameId = values[4] + s.name
		s.gencol = e.GeneratedColumn
		s.source = e.SourceFileId
		s.line = e.SourceLine
		s.column = e.SourceColumn
		s.name = e.NameId
	default:
		err = fmt.Errorf("unexpected number of values in entry: %v", numValues)
	}

	return
}
