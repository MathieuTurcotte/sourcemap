// Copyright (c) 2013 Mathieu Turcotte
// Licensed under the MIT license.

// Program that reads a source map from stdin and finds the original mapping
// for the given line and column numbers in the generated source.
package main

import (
	"flag"
	"fmt"
	"github.com/MathieuTurcotte/sourcemap"
	"os"
)

var line = flag.Int("line", -1, "line number to lookup")
var column = flag.Int("column", -1, "column number to lookup")

func main() {
	flag.Parse()

	sourceMap, err := sourcemap.Read(os.Stdin)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	mapping, err := sourceMap.GetSourceMapping(*line, *column)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	fmt.Printf("%+v\n", mapping)
}
