package pdf

import (
	"fmt"
	"os"

	"github.com/leep-frog/command"
	"github.com/unidoc/unipdf/v3/common/license"
	"github.com/unidoc/unipdf/v3/model"
)

func CLI() *PDF {
	return &PDF{}
}

type PDF struct {
	clientInitialized bool
}

func (pdf *PDF) initializeClient() error {
	if pdf.clientInitialized {
		return nil
	}

	// Make sure to load your metered License API key prior to using the library.
	// If you need a key, you can sign up and create a free one at https://cloud.unidoc.io
	if err := license.SetMeteredKey(os.Getenv(`UNIDOC_LICENSE_API_KEY`)); err != nil {
		return fmt.Errorf("failed to load license: %v", err)
	}

	pdf.clientInitialized = true
	return nil
}

func (*PDF) Load(jsn string) error { return nil }
func (*PDF) Changed() bool         { return false }
func (*PDF) Setup() []string       { return nil }
func (*PDF) Name() string {
	return "gdf"
}

func (pdf *PDF) Node() *command.Node {
	return command.BranchNode(map[string]*command.Node{
		"rotate": command.SerialNodes(
			command.FileNode("inputFile"),
			command.FileNode("outputFile"),
			command.StringNode("direction", command.SimpleCompletor("left", "right", "around")),
			command.ExecutorNode(pdf.cliRotate),
		),
	}, nil, true)
}

// cliRotate is a wrapper around pdf.Rotate that can be used as a CLI executor node.
func (pdf *PDF) cliRotate(output command.Output, data *command.Data) error {
	if err := pdf.initializeClient(); err != nil {
		return output.Stderrf("failed to initialize pdf client: %v", err)
	}
	inputPath := data.String("inputFile")
	outputPath := data.String("outputFile")

	var degrees int64
	switch data.String("direction") {
	case "right":
		degrees = 90
	case "around":
		degrees = 180
	case "left":
		degrees = 270
	}

	if err := pdf.Rotate(degrees, inputPath, outputPath); err != nil {
		return output.Stderrf("failed to rotate pdf: %v", err)
	}
	return nil
}

// Rotate all pages by n degrees
func (pdf *PDF) Rotate(degrees int64, inputPath string, outputPath string) error {
	pdfReader, f, err := model.NewPdfReaderFromFile(inputPath, nil)
	if err != nil {
		return err
	}
	defer f.Close()

	pdfWriter, err := pdfReader.ToWriter(&model.ReaderToWriterOpts{})
	if err != nil {
		return nil
	}

	// Rotate all pages by the provided number of degrees.
	err = pdfWriter.SetRotation(int(degrees))
	if err != nil {
		return nil
	}

	pdfWriter.WriteToFile(outputPath)

	return err
}
