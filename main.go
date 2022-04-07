package main

import (
	"encoding/json"
	"fmt"
	"github.com/VlasovArtem/hob-migration/src/client"
	"github.com/VlasovArtem/hob-migration/src/config"
	"github.com/VlasovArtem/hob-migration/src/migrator"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
)

func main() {
	log.Info().Msg("Starting hob-migration")

	cmdConfig := config.NewCMDConfig()
	cmdConfig.Parse()

	log.Info().Msg(fmt.Sprintf("Config details: \n%s", cmdConfig.String()))

	hobClient := client.NewHobClient(cmdConfig)

	validateRequest(cmdConfig, hobClient)

	requestMigrator := readRequestMigrator(cmdConfig)

	var rollbackOperation []func()

	groupMap, rollbackOperation := migrator.
		NewGroupMigrator(requestMigrator, cmdConfig, hobClient).
		Migrate(rollbackOperation)

	houseMap, rollbackOperation := migrator.
		NewHouseMigrator(requestMigrator, cmdConfig, hobClient, groupMap).
		Migrate(rollbackOperation)

	_, rollbackOperation = migrator.
		NewIncomeMigrator(requestMigrator, hobClient, houseMap, groupMap).
		Migrate(rollbackOperation)

	_, rollbackOperation = migrator.
		NewPaymentMigrator(requestMigrator, cmdConfig, hobClient, houseMap).
		Migrate(rollbackOperation)

	log.Info().Msg("Completed hob-migration")
}

func validateRequest(cmdConfig *config.CMDConfig, hobClient *client.HobClient) {
	err := hobClient.HealthCheck()

	if err != nil {
		log.Fatal().Err(err).Msg("Hob API is not available")
		return
	}

	if !hobClient.UserExists(cmdConfig.UserId) {
		log.Fatal().Msg(fmt.Sprintf("user with %s not found", cmdConfig.UserId))
	}
}

func readRequestMigrator(cmdConfig *config.CMDConfig) (migrator migrator.RequestMigrator) {
	open, err := os.Open(cmdConfig.MigratorFilePath)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to open migrator file %s", cmdConfig.MigratorFilePath)
	}
	bytes, err := ioutil.ReadAll(open)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to read migrator file %s", cmdConfig.MigratorFilePath)
	}
	details := new(map[string]string)
	err = json.Unmarshal(bytes, &details)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to unmarshal migrator file %s", cmdConfig.MigratorFilePath)
	}

	migrator.TypeToPathMap = *details
	return migrator
}
