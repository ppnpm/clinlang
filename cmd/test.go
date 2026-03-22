package main

import (
	"fmt"
	"clinlang/pkg/engine"
)

func main() {
	input := `pt 30M
rx paracetamol 500mg bd po 5d`

	c := engine.ParseString(input)
	soap := engine.FormatSOAP(c)

	fmt.Println(soap)
}
