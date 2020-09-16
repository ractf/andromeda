package node

import "github.com/docker/docker/api/types"

type Config struct {
	BindIp              string           `json:"bindIp"`
	PublicIp            string           `json:"publicIp"`
	ApiIp               string           `json:"apiIp"`
	ApiPort             int              `json:"apiPort"`
	PortMin             int              `json:"portMin"`
	PortMax             int              `json:"portMax"`
	ApiKey              string           `json:"apiKey"`
	DefaultRegistryAuth types.AuthConfig `json:"registryAuth,omitempty"`
	DiscordWebhookUrl   string           `json:"discordWebhookUrl"`
	ConfigPath          string
	RefreshConfig       bool
}

var DefaultConfig = Config{
	BindIp:              "127.0.0.1",
	PublicIp:            "127.0.0.1",
	ApiIp:               "127.0.0.1",
	ApiPort:             6000,
	PortMin:             10000,
	PortMax:             65535,
	ApiKey:              "",
	DefaultRegistryAuth: types.AuthConfig{},
	DiscordWebhookUrl:   "",
}
