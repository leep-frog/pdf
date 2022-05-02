package main

import (
	"os"

	"github.com/leep-frog/command/sourcerer"
	"github.com/leep-frog/pdf"
)

func main() {
	os.Exit(sourcerer.Source(pdf.CLI()))
}
