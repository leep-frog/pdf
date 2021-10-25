package pdf

import (
	"fmt"
	"os"

	"github.com/leep-frog/command"
	"github.com/unidoc/unipdf/v3/common/license"
	"github.com/unidoc/unipdf/v3/model"
)

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
	return "pdf"
}

func (pdf *PDF) Node() *command.Node {
	return command.BranchNode(map[string]*command.Node{
		"rotate": command.SerialNodes(
			command.FileNode("inputFile"),
			command.FileNode("outputFile"),
			command.ExecutorNode(pdf.rotate),
		),
	}, nil, true)
}

func (pdf *PDF) rotate(output command.Output, data *command.Data) error {
	inputPath := data.String("inputFile")
	outputPath := data.String("outputFile")

	if err := rotatePdf(inputPath, outputPath); err != nil {
		return output.Stderrf("failed to rotate pdf: %v", err)
	}
	return nil
}

// Rotate all pages by 90 degrees.
func rotatePdf(inputPath string, outputPath string) error {
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
