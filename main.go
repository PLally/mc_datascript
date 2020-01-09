package main

import (
	"fmt"
	"io/ioutil"
)

func main() {
	// src is the input that we want to tokenize.
	srcb, err := ioutil.ReadFile("example_script")
	if err != nil {
		panic(err)
	}
	src := string(srcb)
	// Initialize the scanner.
	fmt.Println("TEST")
	tokens := LexText(src)
	for _, t := range tokens {
		fmt.Println(t.value)
	}

	p := Parser{
		tokens: tokens,
		alias:  make(map[string]string),
	}

	p.run()
}
