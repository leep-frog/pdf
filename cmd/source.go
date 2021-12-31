package main

import (
	"github.com/leep-frog/command/sourcerer"
	"github.com/leep-frog/pdf"
)

func main() {
	sourcerer.Source(pdf.CLI())
}
