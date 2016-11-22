// Copyright (c) 2013 Mathieu Turcotte
// Licensed under the MIT license.

// This packages implement functions to read the source map format described
// in the "Source Map Revision 3 Proposal" available at http://goo.gl/bcVlcK
package sourcemap

import "fmt"

// Represents the mapping to a line/column/name in the original file.
type OriginalMapping struct {
	File   string // The filename of the original file.
	Line   int    // 1-based line number.
	Column int    // 1-based column number.
	Name   string // The symbol name, if any.
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

// A struct representing the decoded source map.
type SourceMap struct {
	Version        int
	File           string
	SourceRoot     string
	Sources        []string
	SourcesContent []string
	Names          []string
	Mappings       []Line
}

// Given a line and a column number in the generated code, find a mapping in
// the original source. The line and column parameters are 1-based. If no
// mapping can be found for a given line, a mapping on the previous line is
// returned.
func (s *SourceMap) GetSourceMapping(linum, column int) (mapping OriginalMapping, err error) {
	linum--
	column--

	if linum < 0 || linum >= len(s.Mappings) {
		err = fmt.Errorf("invalid line number: %v", linum+1)
		return
	}

	if column < 0 {
		err = fmt.Errorf("invalid column number: %v", column+1)
		return
	}

	line := s.Mappings[linum]

	if len(line) == 0 || line[0].GeneratedColumn > column {
		return s.getPreviousLineMapping(linum)
	}

	index := -1
	for i, entry := range line {
		if entry.GeneratedColumn <= column {
			index = i
			break
		}
	}

	if index < 0 {
		err = fmt.Errorf("unable to map column: %d", column+1)
		return
	}

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

		line := s.Mappings[linum]

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
