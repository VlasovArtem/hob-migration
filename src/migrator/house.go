package migrator

import (
	"fmt"
	"github.com/VlasovArtem/hob-migration/src/client"
	"github.com/VlasovArtem/hob-migration/src/config"
	"github.com/VlasovArtem/hob-migration/src/model"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"strings"
)

type HouseMigrator struct {
	*BaseMigrator[map[string]model.HouseDto]
	client   *client.HobClient
	groupMap map[string]model.GroupDto
	config   *config.CMDConfig
}

func NewHouseMigrator(
	requestMigrator RequestMigrator,
	config *config.CMDConfig,
	hobClient *client.HobClient,
	groupMap map[string]model.GroupDto,
) *HouseMigrator {
	log.Info().Msg("Starting House Migrator")

	path, ok := requestMigrator.TypeToPathMap["houses"]
	if !ok {
		log.Info().Msg("houses path not found")
		return nil
	}
	migrator := &HouseMigrator{
		client:   hobClient,
		groupMap: groupMap,
		config:   config,
	}
	filePath := path
	migrator.BaseMigrator = &BaseMigrator[map[string]model.HouseDto]{
		mappers: map[string]Mapper[map[string]model.HouseDto]{
			"csv": &CSVMigrator[MapCreateHouseRequest, map[string]model.HouseDto]{
				filePath: filePath,
				header:   []string{"House Identifier", "Groups", "Name", "Country", "City", "Address 1", "Address 2"},
				parser:   migrator.parseCSVLine(),
				mapper:   migrator.mapHouses,
			},
		},
		filePath: filePath,
	}

	return migrator
}

func (h *HouseMigrator) mapHouses(requests []MapCreateHouseRequest) (map[string]model.HouseDto, error) {
	var response = make(map[string]model.HouseDto)

	for _, request := range requests {
		house, err := h.client.CreateHouse(request.request)
		if err != nil {
			log.Fatal().Err(err)
		} else {
			response[request.identifier] = house
		}
	}

	log.Info().Msg(fmt.Sprintf("%d houses created", len(response)))

	return response, nil
}

func (h *HouseMigrator) parseCSVLine() func(line []string, lineNumber int) (MapCreateHouseRequest, error) {
	return func(line []string, lineNumber int) (MapCreateHouseRequest, error) {
		var groupIds []uuid.UUID

		for _, groupName := range strings.Split(line[1], ",") {
			if dto, ok := h.groupMap[groupName]; !ok {
				log.Fatal().Msgf("group with name %s not found at the csv line %d", groupName, lineNumber)
			} else {
				groupIds = append(groupIds, dto.Id)
			}
		}
		request := model.CreateHouseRequest{
			GroupIds:    groupIds,
			Name:        line[2],
			CountryCode: line[3],
			City:        line[4],
			StreetLine1: line[5],
			StreetLine2: line[6],
			UserId:      h.config.UserId,
		}

		return MapCreateHouseRequest{
			identifier: line[0],
			request:    request,
		}, nil
	}
}

type MapCreateHouseRequest struct {
	identifier string
	request    model.CreateHouseRequest
}
