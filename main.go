package main

import (
	"encoding/json"
	"fmt"
	"github.com/VlasovArtem/hob-migration/src/client"
	"github.com/VlasovArtem/hob-migration/src/config"
	"github.com/VlasovArtem/hob-migration/src/migrator"
	"github.com/VlasovArtem/hob-migration/src/model"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
)

func main() {
	cmdConfig := config.NewCMDConfig()
	cmdConfig.Parse()

	hobClient := client.NewHobClient(cmdConfig)

	validateRequest(cmdConfig, hobClient)

	requestMigrator := readRequestMigrator(cmdConfig)

	groupMap := migrator.Migrate[map[string]model.GroupDto](migrator.NewGroupMigrator(requestMigrator, hobClient).BaseMigrator)
	houseMap := migrator.Migrate[map[string]model.HouseDto](migrator.NewHouseMigrator(requestMigrator, cmdConfig, hobClient, groupMap).BaseMigrator)
	migrator.Migrate[[]model.IncomeDto](migrator.NewIncomeMigrator(requestMigrator, hobClient, houseMap, groupMap).BaseMigrator)
	migrator.Migrate[[]model.PaymentDto](migrator.NewPaymentMigrator(requestMigrator, cmdConfig, hobClient, houseMap).BaseMigrator)
}

func validateRequest(cmdConfig *config.CMDConfig, hobClient *client.HobClient) {
	err := hobClient.HealthCheck()

	if err != nil {
		log.Fatal().Err(err)
	}

	if !hobClient.UserExists(cmdConfig.UserId) {
		log.Fatal().Msg(fmt.Sprintf("user with %s not found", cmdConfig.UserId))
	}
}

func readRequestMigrator(cmdConfig *config.CMDConfig) (migrator migrator.RequestMigrator) {
	open, err := os.Open(cmdConfig.MigratorFilePath)
	if err != nil {
		log.Fatal().Err(err)
	}
	bytes, err := ioutil.ReadAll(open)
	if err != nil {
		log.Fatal().Err(err)
	}
	details := new(map[string]string)
	err = json.Unmarshal(bytes, &details)
	if err != nil {
		log.Fatal().Err(err)
	}

	migrator.TypeToPathMap = *details
	return migrator
}
