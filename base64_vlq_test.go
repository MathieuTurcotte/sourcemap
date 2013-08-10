// Copyright (c) 2013 Mathieu Turcotte
// Licensed under the MIT license.

package sourcemap

import (
	"strings"
	"testing"
)

func TestDecodeVLQ(t *testing.T) {
	cases := []struct {
		input   string
		results []int
	}{
		// Simple positive VLQ encoded values.
		{"A", []int{0}},  // 000000
		{"B", []int{0}},  // 000001
		{"C", []int{1}},  // 000010
		{"D", []int{-1}}, // 000011
		// Positive VQL encoded values with a continuation bit.
		{"gA", []int{0}},  // 100000 000000
		{"gB", []int{16}}, // 100000 000001
		{"gC", []int{32}}, // 100000 000010
		{"gD", []int{48}}, // 100000 000011
		// Negative VLQ encoded values with a continuation bit.
		{"hA", []int{0}},   // 100001 000000
		{"hB", []int{-16}}, // 100001 000001
		{"hC", []int{-32}}, // 100001 000010
		{"hD", []int{-48}}, // 100001 000011
		// Consecutive VLQ encoded values.
		{"AA", []int{0, 0}},   // 000000 000000
		{"BB", []int{0, 0}},   // 000001 000001
		{"CC", []int{1, 1}},   // 000010 000010
		{"DD", []int{-1, -1}}, // 000011 000011
	}

	for _, c := range cases {
		reader := strings.NewReader(c.input)
		for _, r := range c.results {
			val, err := decodeVQL(reader)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			} else if val != r {
				t.Errorf("expected %v as decoded value, got %v", r, val)
			}
		}

		// Ensure that all the bytes were consumed from the reader.
		if reader.Len() != 0 {
			t.Errorf("did not read input fully %+v", c)
		}
	}

	for _, c := range []string{"", "g", "gg"} {
		if _, err := decodeVQL(strings.NewReader(c)); err == nil {
			t.Errorf("expected error for %v, got nil", c)
		}
	}
}
