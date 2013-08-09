package main

import (
	"flag"
	"fmt"
	"github.com/MathieuTurcotte/sourcemap"
	"os"
)

var line = flag.Int("line", -1, "line number to lookup")
var column = flag.Int("column", -1, "column number to lookup")
var printMap = flag.Bool("print", false, "whether to print the source map")

func main() {
	flag.Parse()

	sourceMap, err := sourcemap.Read(os.Stdin)

	if err != nil {
		fmt.Println(err)
	} else if *printMap {
		fmt.Printf("%+v\n", sourceMap)
	}

	mapping, err := sourceMap.GetSourceMapping(*line, *column)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%+v\n", mapping)
	}
}
