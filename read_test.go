// Copyright (c) 2013 Mathieu Turcotte
// Licensed under the MIT license.

package sourcemap

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestRead(t *testing.T) {
	cases := []struct {
		sourceMapPath           string
		serializedSourceMapPath string
	}{
		{"testdata/sample.sourcemap", "testdata/sample.json"},
	}

	for _, c := range cases {
		testRead(t, c.sourceMapPath, c.serializedSourceMapPath)
	}
}

func testRead(t *testing.T, sourceMapPath, serializedSourceMapPath string) {
	var decodedSourceMap SourceMap
	var serializedSourceMap SourceMap

	// Read the serialized source map.
	serializedSourceMapFile, err := os.Open(serializedSourceMapPath)
	if err != nil {
		t.Fatalf("could not open serialized source map: %s", err)
	}
	defer serializedSourceMapFile.Close()

	dec := json.NewDecoder(serializedSourceMapFile)
	if err := dec.Decode(&serializedSourceMap); err != nil {
		t.Fatalf("could not open serialized source map: %s", err)
	}

	// Read the source map to decode.
	sourceMapFile, err := os.Open(sourceMapPath)
	if err != nil {
		t.Fatalf("could not open serialized source map: %s", err)
	}
	defer sourceMapFile.Close()

	decodedSourceMap, err = Read(sourceMapFile)
	if err != nil {
		t.Errorf("unexpected error when decoding source map: %s", err)
	}

	if !reflect.DeepEqual(decodedSourceMap, serializedSourceMap) {
		t.Errorf("decoded source map doesn't match serialized source map")
		t.Errorf("decoded source map: %+v", decodedSourceMap)
		t.Errorf("serialized source map: %+v", serializedSourceMap)
	}
}
