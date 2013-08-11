// Copyright (c) 2013 Mathieu Turcotte
// Licensed under the MIT license.

package sourcemap

import (
	"encoding/json"
	"os"
	"testing"
)

func readSerializedSourceMap(path string) (s SourceMap, err error) {
	f, err := os.Open("testdata/sample.json")
	if err != nil {
		return
	}

	dec := json.NewDecoder(f)
	err = dec.Decode(&s)

	return
}

func TestGetSourceMapping(t *testing.T) {
	sourceMap, err := readSerializedSourceMap("testdata/sample.json")
	if err != nil {
		t.Fatalf("could not open test data, got error %s", err)
	}

	cases := []struct {
		line    int
		column  int
		mapping OriginalMapping
	}{
		// 6: number = 42
		{6, 3, OriginalMapping{"example.coffee", 2, 1, "number"}},
		// 8: opposite = true
		{8, 3, OriginalMapping{"example.coffee", 3, 1, "opposite"}},
		// 10: if (opposite) {
		{10, 3, OriginalMapping{"example.coffee", 6, 1, ""}},
		{11, 5, OriginalMapping{"example.coffee", 6, 1, ""}},
		// 14: square = function(x) {
		{14, 3, OriginalMapping{"example.coffee", 9, 1, "square"}},
		{15, 5, OriginalMapping{"example.coffee", 9, 10, ""}},
		// 18: list = [1, 2, 3, 4, 5];
		{18, 3, OriginalMapping{"example.coffee", 12, 1, "list"}},
	}

	for _, c := range cases {
		mapping, err := sourceMap.GetSourceMapping(c.line, c.column)
		if err != nil {
			t.Errorf("unexpected error %v", err)
		} else if c.mapping != mapping {
			t.Errorf("expected mapping %+v for (%v, %v), got %+v", c.mapping,
				c.line, c.column, mapping)
		}
	}
}
