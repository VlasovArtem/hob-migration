package config

import (
	"fmt"
	"github.com/spf13/pflag"
)

type CMDConfig struct {
	HobURL           string
	MigratorFilePath string
	UserId           string
}

func NewCMDConfig() *CMDConfig {
	return &CMDConfig{}
}

func (c *CMDConfig) Parse() {
	pflag.StringVarP(&c.HobURL, "url", "u", "http://localhost:3030", "URL to HOB application.")
	pflag.StringVarP(&c.MigratorFilePath, "migrator-path", "m", "", fmt.Sprintf("Path to the migrator file path. Details:\n%s)", migrationDetails()))
	pflag.StringVarP(&c.UserId, "user-id", "i", "", "User id")
	pflag.Parse()
}

func migrationDetails() string {
	return "Example of the migrator json:\n{\"groups\":\"path_to_the_file\"}\n\nPossible Values of keys:\n- groups\n- houses\n- incomes\n- payments"
}
