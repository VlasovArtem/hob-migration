package parser

import (
	"encoding/csv"
	"github.com/VlasovArtem/hob-migration/src/validator"
	"os"
)

func Parse[T any](path string, header []string, parser func(line []string) (T, error)) ([]T, error) {
	open, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer open.Close()

	csvReader := csv.NewReader(open)
	data, err := csvReader.ReadAll()

	if err != nil {
		return nil, err
	}

	var items []T

	for i, line := range data {
		if i == 0 {
			if err := validator.VerifyCSVHeader(header, line); err != nil {
				return nil, err
			}
		} else {
			if item, err := parser(line); err != nil {
				return nil, err
			} else {
				items = append(items, item)
			}
		}
	}

	return items, open.Close()
}
