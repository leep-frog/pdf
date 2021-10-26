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
	key := os.Getenv(`UNIDOC_LICENSE_API_KEY`)
	if err := license.SetMeteredKey(key); err != nil {
		return fmt.Errorf("failed to load license (%s): %v", key, err)
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

	if err := pdf.Rotate(inputPath, outputPath); err != nil {
		return output.Stderrf("failed to rotate pdf: %v", err)
	}
	return nil
}

// Rotate all pages by 90 degrees.
func (pdf *PDF) Rotate(inputPath string, outputPath string) error {
	pdfReader, f, err := model.NewPdfReaderFromFile(inputPath, nil)
	if err != nil {
		return err
	}
	defer f.Close()

	pdfWriter, err := pdfReader.ToWriter(&model.ReaderToWriterOpts{})
	if err != nil {
		return nil
	}

	// Rotate all page 90 degrees.
	err = pdfWriter.SetRotation(90)
	if err != nil {
		return nil
	}

	pdfWriter.WriteToFile(outputPath)

	return err
}
