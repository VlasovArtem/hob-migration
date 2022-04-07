package parser

import (
	"encoding/csv"
	"github.com/VlasovArtem/hob-migration/src/validator"
	"github.com/rs/zerolog/log"
	"os"
)

func Parse[T any](path string, header []string, parser func(line []string, lineNumber int) (T, error)) ([]T, error) {
	open, err := os.Open(path)

	if err != nil {
		log.Error().Err(err).Msgf("Can't open file %s", path)
		return nil, err
	}

	defer open.Close()

	csvReader := csv.NewReader(open)
	data, err := csvReader.ReadAll()

	if err != nil {
		log.Error().Err(err).Msgf("Can't read file %s", path)
		return nil, err
	}

	var items []T

	log.Info().Msgf("Start parsing %s", path)

	for i, line := range data {
		if i == 0 {
			if err := validator.VerifyCSVHeader(header, line); err != nil {
				log.Err(err).Msgf("Can't parse file %s", path)
				return nil, err
			}
		} else {
			if item, err := parser(line, i); err != nil {
				return nil, err
			} else {
				items = append(items, item)
			}
		}
	}

	return items, open.Close()
}
