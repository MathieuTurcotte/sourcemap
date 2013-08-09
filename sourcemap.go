// This packages implement functions to read the source map format described
// in the "Source Map Revision 3 Proposal" available at http://goo.gl/bcVlcK
package sourcemap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

const maxEntryValue = 5    // Maximum number of values within a single entry.
const lineSeparator = ';'  // Line separator in the "mappings" field.
const entrySeparator = ',' // Entry separator withing a line in the "mappings" field.

// Represents the mapping to a line/column/name in the original file.
type OriginalMapping struct {
	File   string
	Line   int
	Column int
	Name   string
}

// A struct into which the source map data is unmarshalled.
type SourceMap struct {
	Version        int
	File           string
	SourceRoot     string
	Sources        []string
	SourcesContent []string
	Names          []string
	Mappings       Mappings
}

// Given a line and a column number in the generated code, find a mapping in
// the original source. The line and column parameters are 1-based. If no
// mapping can be found for a given line, a mapping on the previous line is
// returned.
func (s *SourceMap) GetSourceMapping(linum, column int) (mapping OriginalMapping, err error) {
	linum--
	column--

	if linum < 0 || linum > len(s.Mappings.Lines) {
		err = fmt.Errorf("invalid line number: %v", linum+1)
		return
	}

	if column < 0 {
		err = fmt.Errorf("invalid column number: %v", column+1)
		return
	}

	line := s.Mappings.Lines[linum]

	if len(line) == 0 || line[0].GeneratedColumn > column {
		return s.getPreviousLineMapping(linum)
	}

	index := sort.Search(len(line), func(i int) bool {
		return line[i].GeneratedColumn <= column
	})

	entry := line[index]
	s.populateMapping(&mapping, entry)

	return
}

func (s *SourceMap) getPreviousLineMapping(linum int) (mapping OriginalMapping, err error) {
	for {
		linum--

		if linum < 0 {
			err = fmt.Errorf("cannot find previous line mapping")
			return
		}

		line := s.Mappings.Lines[linum]

		if len(line) > 0 {
			entry := line[len(line)-1]
			s.populateMapping(&mapping, entry)
			return
		}
	}
}

func (s *SourceMap) populateMapping(mapping *OriginalMapping, entry Entry) {
	mapping.File = s.Sources[entry.SourceFileId]
	mapping.Line = entry.SourceLine + 1
	mapping.Column = entry.SourceColumn + 1

	if entry.NameId >= 0 {
		mapping.Name = s.Names[entry.NameId]
	}
}

// Represents a section of the generated source that can be mapped back to the
// original source.
type Entry struct {
	GeneratedColumn int // 0-based column of this entry in the generated source.
	SourceFileId    int // Index of the source file in the "sources" list.
	SourceLine      int // 0-based line number of this entry in the source file.
	SourceColumn    int // 0-based column number of this entry in the source file.
	NameId          int // Index of the symbol name in the "names" list.
}

// Represents a line in the generated source. A line is composed of entries
// containing information about the original source file.
type Line []Entry

// The parsed representation of a source map mapping. Slots in the lines slice
// will be null if the line does not have any entries.
type Mappings struct {
	Lines []Line
}

// Struct to hold the state while parsing the "mappings" field.
type unmarshalState struct {
	gencol int // 0-based index of the generated column.
	source int // 0-based index into the "sources" list.
	line   int // 0-based starting line in the original source.
	column int // 0-based starting column of the line in the source.
	name   int // 0-based index into the "names" list.
}

func (m *Mappings) UnmarshalJSON(data []byte) error {
	m.Lines = make([]Line, 0, 256)

	// Instead of splitting the input string into multiple
	reader := bytes.NewReader(data)

	state := unmarshalState{}
	line := make(Line, 0, 10)
	for reader.Len() > 0 {
		if consume(reader, lineSeparator) {
			m.Lines = append(m.Lines, line)
			line = make(Line, 0, 10)
			state.gencol = 0
		} else {
			i := 0
			var values [maxEntryValue]int
			for !entryCompleted(reader) {
				if val, err := decodeVQL(reader); err != nil {
					return err
				} else {
					values[i] = val
					i++
				}
			}

			entry, err := newEntry(&state, values, i)
			if err != nil {
				return err
			}

			line = append(line, entry)
			consume(reader, entrySeparator)
		}
	}

	return nil
}

func consume(r *bytes.Reader, wanted byte) bool {
	if r.Len() > 0 {
		b, _ := r.ReadByte()
		if b == wanted {
			return true
		}
		r.UnreadByte()
	}
	return false
}

func entryCompleted(r *bytes.Reader) bool {
	if r.Len() > 0 {
		b, _ := r.ReadByte()
		r.UnreadByte()
		return b == lineSeparator || b == entrySeparator
	}
	return true
}

func newEntry(s *unmarshalState, values [maxEntryValue]int,
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

// Reads a source map.
func Read(reader io.Reader) (s SourceMap, err error) {
	dec := json.NewDecoder(reader)
	if err = dec.Decode(&s); err != nil {
		return
	}

	if s.Version != 3 {
		err = fmt.Errorf("unsupported version: %v", s.Version)
	}

	return
}
