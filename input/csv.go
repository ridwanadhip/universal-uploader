package input

import (
	"encoding/csv"
	"io"
	"os"
)

type csvParser struct {
	batchSize    int
	currentIndex int
	file         *os.File
	reader       *csv.Reader
}

func (parser *csvParser) NextBatch(path string) (batch *Batch, exists bool, err error) {
	reader := parser.reader
	if reader == nil {
		f, err := os.Open(path)
		if err != nil {
			return nil, false, err
		}

		parser.file = f
		parser.reader = csv.NewReader(f)
		reader = parser.reader
	}

	data := [][]string{}
	for {
		columns, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, false, err
		}

		parser.currentIndex += 1

		// skip header
		if parser.currentIndex == 1 {
			continue
		}

		data = append(data, columns)
		if len(data) == parser.batchSize {
			break
		}
	}

	batch = &Batch{
		Data:  data,
		Index: parser.currentIndex - len(data) - 1,
	}

	// end of file
	if len(batch.Data) == 0 {
		parser.reset()
		return nil, false, nil
	}

	return batch, true, nil
}

func (parser *csvParser) GetFieldNames(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	csvReader := csv.NewReader(f)
	header, err := csvReader.Read()

	return header, err
}

func (parser *csvParser) GetCurrentIndex() int {
	return parser.currentIndex
}

func (parser *csvParser) Close() {
	if parser.reader != nil {
		parser.file.Close()
		parser.reset()
	}
}

func (parser *csvParser) reset() {
	parser.file = nil
	parser.reader = nil
	parser.currentIndex = 0
}
