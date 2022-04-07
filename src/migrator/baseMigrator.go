package migrator

import (
	"github.com/VlasovArtem/hob-migration/src/parser"
	"github.com/VlasovArtem/hob-migration/src/validator"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"strings"
)

type RequestMigrator struct {
	TypeToPathMap map[string]string
}

type Migrator[RESPONSE any] interface {
	Migrate() (RESPONSE, error)
}

type Mapper[RESPONSE any] interface {
	Map() (RESPONSE, error)
}

type BaseMigrator[RESPONSE any] struct {
	mappers  map[string]Mapper[RESPONSE]
	filePath string
	Rollback func()
}

func (b *BaseMigrator[RESPONSE]) Migrate(rollbackOnError []func()) RESPONSE {
	if b == nil {
		return *new(RESPONSE)
	}

	if err := b.Verify(); err != nil {
		rollback(rollbackOnError)
	}

	t, err := b.mappers[strings.Replace(filepath.Ext(b.filePath), ".", "", 1)].Map()

	if err != nil {
		log.Error().Err(err).Msg("Error while migrating")
		rollback(rollbackOnError)
	}

	return t
}

func (b *BaseMigrator[T]) Verify() error {
	return validator.Validate(
		func() error {
			return validator.VerifyFilePathIsEmpty(b.filePath, "apartments file path is empty")
		},
		func() error {
			return validator.VerifyFilePathTypeIsValid(b.filePath)
		},
		func() error {
			return validator.VerifyFilePathExists(b.filePath)
		},
	)
}

func rollback(rollbackOnError []func()) {
	if len(rollbackOnError) != 0 {
		for i := len(rollbackOnError) - 1; i >= 0; i-- {
			rollbackOnError[i]()
		}
	}

	os.Exit(1)
}

type CSVMigrator[REQUEST any, RESPONSE any] struct {
	filePath string
	header   []string
	parser   func(line []string, lineNumber int) (REQUEST, error)
	mapper   func(requests []REQUEST) (RESPONSE, error)
}

func (c *CSVMigrator[REQUEST, RESPONSE]) Map() (response RESPONSE, err error) {
	log.Info().Msgf("Start CSV Migration for file: %s", c.filePath)

	requests, err := parser.Parse[REQUEST](c.filePath, c.header, c.parser)

	if err != nil {
		log.Error().Err(err).Msgf("Error while parsing CSV file")
		return response, err
	}

	return c.mapper(requests)
}
