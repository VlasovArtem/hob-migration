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
	log.Info().Msg("Starting hob-migration")

	cmdConfig := config.NewCMDConfig()
	cmdConfig.Parse()

	log.Info().Msg(fmt.Sprintf("Config details: \n%s", cmdConfig.String()))

	hobClient := client.NewHobClient(cmdConfig)

	validateRequest(cmdConfig, hobClient)

	requestMigrator := readRequestMigrator(cmdConfig)

	var rollbackOperation []func()

	groupMigrator := migrator.NewGroupMigrator(requestMigrator, cmdConfig, hobClient)
	groupMap := groupMigrator.Migrate[map[string]model.GroupDto](rollbackOperation)

	rollbackOperation = append(rollbackOperation, func() { groupMigrator.Rollback(groupMap) })

	houseMigrator := migrator.NewHouseMigrator(requestMigrator, cmdConfig, hobClient, groupMap)
	houseMap := houseMigrator.Migrate[map[string]model.HouseDto](rollbackOperation)

	rollbackOperation = append(rollbackOperation, func() { houseMigrator.Rollback(houseMap) })

	incomeMigrator := migrator.NewIncomeMigrator(requestMigrator, hobClient, houseMap, groupMap)
	incomes := incomeMigrator.Migrate[[]model.IncomeDto](rollbackOperation)

	rollbackOperation = append(rollbackOperation, func() { incomeMigrator.Rollback(incomes) })

	paymentMigrator := migrator.NewPaymentMigrator(requestMigrator, cmdConfig, hobClient, houseMap)
	payments := paymentMigrator.Migrate[[]model.PaymentDto](rollbackOperation)

	rollbackOperation = append(rollbackOperation, func() { paymentMigrator.Rollback(payments) })

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
