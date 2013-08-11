// Copyright (c) 2013 Mathieu Turcotte
// Licensed under the MIT license.

// Program that reads a source map from stdin and outputs a JSON representation
// of the decoded source map. Used to generate testing source maps that can be
// read into test cases.
package main

import (
	"encoding/json"
	"fmt"
	"github.com/MathieuTurcotte/sourcemap"
	"os"
)

func main() {
	sourceMap, err := sourcemap.Read(os.Stdin)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	b, err := json.MarshalIndent(sourceMap, "", "  ")
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	_, err = os.Stdout.Write(b)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}
