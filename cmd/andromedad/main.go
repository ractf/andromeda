package main

import (
	"encoding/json"
	"fmt"
	"github.com/ractf/andromeda/pkg/api"
	"github.com/ractf/andromeda/pkg/node"
	"github.com/spf13/cobra"
	"io/ioutil"
	"math/rand"
	"os"
	"time"
)

var RootCommand = &cobra.Command{
	Use: "andromedad",
	Run: func(cmd *cobra.Command, args []string) {
		rand.Seed(time.Now().Unix())
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

		node := node.StartNode(&config)
		apiServer := api.Server{
			Node: node,
		}

		err = (&apiServer).StartAPIServer(&config)
		if err != nil {
			fmt.Println(err)
		}
	},
}

var configPath string

func init() {
	RootCommand.Flags().StringVarP(&configPath, "config_file", "c", "/etc/andromeda/config.json", "config_file")
}

func main() {
	err := RootCommand.Execute()
	if err != nil {
		panic(err)
	}
}
