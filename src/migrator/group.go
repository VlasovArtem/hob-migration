package migrator

import (
	"fmt"
	"github.com/VlasovArtem/hob-migration/src/client"
	"github.com/VlasovArtem/hob-migration/src/config"
	"github.com/VlasovArtem/hob-migration/src/model"
	"github.com/rs/zerolog/log"
)

type GroupMigrator struct {
	*BaseMigrator[map[string]model.GroupDto]
	config *config.CMDConfig
	client *client.HobClient
}

func NewGroupMigrator(requestMigrator RequestMigrator, config *config.CMDConfig, hobClient *client.HobClient) *GroupMigrator {
	log.Info().Msg("Starting Group Migrator")

	path, ok := requestMigrator.TypeToPathMap["groups"]
	if !ok {
		log.Info().Msg("groups path not found")
		return nil
	}
	migrator := &GroupMigrator{
		client: hobClient,
		config: config,
	}
	filePath := path
	migrator.BaseMigrator = &BaseMigrator[map[string]model.GroupDto]{
		mappers: map[string]Mapper[map[string]model.GroupDto]{
			"csv": &CSVMigrator[model.CreateGroupRequest, map[string]model.GroupDto]{
				filePath: filePath,
				header:   []string{"Name"},
				parser:   migrator.parseCSVLine(),
				mapper:   migrator.mapGroups,
			},
		},
		filePath: filePath,
	}

	return migrator
}

func (g *GroupMigrator) mapGroups(requests []model.CreateGroupRequest) (map[string]model.GroupDto, error) {
	var response = make(map[string]model.GroupDto)

	if batchResponse, err := g.client.CreateGroupBatch(model.CreateGroupBatchRequest{Groups: requests}); err != nil {
		return nil, err
	} else {
		for _, group := range batchResponse {
			response[group.Name] = group
		}

		log.Info().Msg(fmt.Sprintf("%d groups created", len(response)))
	}

	return response, nil
}

func (g *GroupMigrator) parseCSVLine() func(line []string, lineNumber int) (model.CreateGroupRequest, error) {
	return func(line []string, lineNumber int) (model.CreateGroupRequest, error) {
		request := model.CreateGroupRequest{
			Name:    line[0],
			OwnerId: g.config.UserId,
		}

		return request, nil
	}
}

func (g *GroupMigrator) Rollback(data map[string]model.GroupDto) {
	log.Info().Msg("Rolling back groups")
	if len(data) == 0 {
		log.Info().Msg("No groups to rollback")
	}

	for _, group := range data {
		if err := g.client.DeleteGroupById(group.Id); err != nil {
			log.Error().Err(err).Msgf("Failed to delete group with id %s and name %s", group.Id, group.Name)
		} else {
			log.Info().Msgf("Group with id %s and name %s deleted", group.Id, group.Name)
		}
	}
}
