package pdf

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

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
	input := command.FileNode("inputFile")
	output := command.FileNode("outputFile")
	return command.BranchNode(map[string]*command.Node{
		"rotate": command.SerialNodes(
			input, output,
			command.StringNode("direction", command.SimpleCompletor("left", "right", "around")),
			command.ExecutorNode(pdf.cliRotate),
		),
		"crop": command.SerialNodes(
			input, output,
			command.NewFlagNode(
				command.BoolFlag("landscape", 'l'),
			),
			command.StringNode("paperSize"),
			command.ExecutorNode(pdf.cliCrop),
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
	// TODO: command package: argument type oneof;
	// provide map[value]func? or just []value?
	switch data.String("direction") {
	case "right":
		degrees = 90
	case "around":
		degrees = 180
	case "left":
		degrees = 270
	default:
		return output.Stderr("direction must be either right, left, or around")
	}

	if err := pdf.Rotate(degrees, inputPath, outputPath); err != nil {
		return output.Stderrf("failed to rotate pdf: %v", err)
	}
	return nil
}

// cliCrop is a wrapper around pdf.Crop that can be used as a CLI executor node.
func (pdf *PDF) cliCrop(output command.Output, data *command.Data) error {
	if err := pdf.initializeClient(); err != nil {
		return output.Stderrf("failed to initialize pdf client: %v", err)
	}
	inputPath := data.String("inputFile")
	outputPath := data.String("outputFile")

	dimensions, err := paperSize(data.String("paperSize"))
	if err != nil {
		return output.Err(err)
	}

	if err := pdf.Crop(dimensions[0], dimensions[1], inputPath, outputPath); err != nil {
		return output.Stderrf("failed to crop pdf: %v", err)
	}
	return nil
}

var (
	zeroSizes = map[string][]float64{
		"a": {33.1, 46.8},
		"b": {39.4, 55.7},
	}
	keywordSizes = map[string][]float64{
		"letter": {8.5, 11},
	}
	codeRegex = regexp.MustCompile("^([ab])([0-9])$")
)

func paperSize(code string) ([]float64, error) {
	if size, ok := keywordSizes[code]; ok {
		return size, nil
	}

	m := codeRegex.FindStringSubmatch(strings.ToLower(code))
	if len(m) == 0 {
		return nil, fmt.Errorf("invalid paper code: %q", code)
	}

	letter := m[1]
	index, err := strconv.Atoi(m[2])
	if err != nil {
		return nil, fmt.Errorf("failed to convert string to int: %v", err)
	}

	size, ok := zeroSizes[letter]
	if !ok {
		return nil, fmt.Errorf("invalid paper code: %v", err)
	}

	for i := 0; i < index; i++ {
		width := size[0]
		size[0] = size[1] / 2.0
		size[1] = width
	}

	// PDF units = 1/72 inches so convert to units.
	size[0] *= 72
	size[1] *= 72

	return size, nil
}

/*

Command usage:
  Run blaze commands
  z *^

    test blaze targets
    t *^ TARGET [FUNCTION ...] --abc ABC ABC ... --other

  	build blaze targets
  	b *^ TARGET

  	run blaze targets
  	r *^ TARGET

Symbols:
  *: start of aliasable command
	^: start of cachable command

Arguments:
  FUNCTION: ...
  TARGET: ...

Flags:
  abc: ...
	other: ...

*/

// Example: https://unidoc.io/unipdf-examples/crop-page-content-pdf/
func (pdf *PDF) Crop(width, height float64, inputPath string, outputPath string) error {
	pdfReader, f, err := model.NewPdfReaderFromFile(inputPath, nil)
	if err != nil {
		return err
	}
	defer f.Close()

	opts := &model.ReaderToWriterOpts{
		PageProcessCallback: func(pageNum int, page *model.PdfPage) error {
			bbox, err := page.GetMediaBox()
			if err != nil {
				return err
			}

			// Crop from top left corner, so we only change lower left y (lly) and upper right x (urx).
			bbox.Lly = bbox.Ury - height
			bbox.Urx = bbox.Llx + width

			page.MediaBox = bbox
			return nil
		},
	}

	// Generate a PdfWriter instance from existing PdfReader.
	pdfWriter, err := pdfReader.ToWriter(opts)
	if err != nil {
		return err
	}

	// Write to file.
	return pdfWriter.WriteToFile(outputPath)
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

	f.Close()

	// Rotate all pages by the provided number of degrees.
	err = pdfWriter.SetRotation(degrees)
	if err != nil {
		return nil
	}

	return pdfWriter.WriteToFile(outputPath)
}
