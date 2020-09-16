package main

import (
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/ractf/andromeda/pkg/api"
	"github.com/ractf/andromeda/pkg/node"
	"github.com/spf13/cobra"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"time"
)

var RootCommand = &cobra.Command{
	Use: "andromedad",
	Run: func(cmd *cobra.Command, args []string) {
		rand.Seed(time.Now().Unix())
		var config *node.Config
		if envConfig {
			config = loadConfigFromEnvironment()
		} else {
			config = loadConfigFromFile()
		}

		node := node.StartNode(config)
		apiServer := api.Server{
			Node: node,
		}

		err := (&apiServer).StartAPIServer(config)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func loadConfigFromFile() *node.Config {
	configFile, err := os.Open(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}

		bytes, _ := json.MarshalIndent(node.DefaultConfig, "", "  ")
		err = ioutil.WriteFile(configPath, bytes, 0644)

		configFile, err = os.Open(configPath)
		if err != nil {
			panic(err)
		}
	}

	bytes, err := ioutil.ReadAll(configFile)
	if err != nil {
		panic(err)
	}

	var config node.Config
	err = json.Unmarshal(bytes, &config)
	config.ConfigPath = configPath

	return &config
}

func loadConfigFromEnvironment() *node.Config {
	apiPort, err := strconv.Atoi(os.Getenv("ANDROMEDA_API_PORT"))
	if err != nil {
		panic(err)
	}
	portMin, err := strconv.Atoi(os.Getenv("ANDROMEDA_PORT_MIN"))
	if err != nil {
		panic(err)
	}
	portMax, err := strconv.Atoi(os.Getenv("ANDROMEDA_PORT_MAX"))
	if err != nil {
		panic(err)
	}
	return &node.Config{
		BindIp:   os.Getenv("ANDROMEDA_BIND_IP"),
		PublicIp: os.Getenv("ANDROMEDA_PUBLIC_IP"),
		ApiIp:    os.Getenv("ANDROMEDA_API_IP"),
		ApiPort:  apiPort,
		PortMin:  portMin,
		PortMax:  portMax,
		ApiKey:   os.Getenv("ANDROMEDA_API_KEY"),
		DefaultRegistryAuth: types.AuthConfig{
			Username: os.Getenv("ANDROMEDA_REGISTRY_USERNAME"),
			Password: os.Getenv("ANDROMEDA_REGISTRY_PASSWORD"),
		},
		DiscordWebhookUrl: os.Getenv("ANDROMEDA_DISCORD_WEBHOOK_URL"),
		ConfigPath:        "",
		RefreshConfig:     false,
	}
}

var configPath string
var envConfig bool

func init() {
	RootCommand.Flags().StringVarP(&configPath, "config_file", "c", "/etc/andromeda/config.json", "config_file")
	RootCommand.Flags().BoolVarP(&envConfig, "env_config", "e", false, "env_config")
}

func main() {
	err := RootCommand.Execute()
	if err != nil {
		panic(err)
	}
}
