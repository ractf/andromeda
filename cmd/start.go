package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"math/rand"
	"ractf.co.uk/andromeda/api"
	"ractf.co.uk/andromeda/challenge"
	"time"
)

var StartCommand = &cobra.Command{
	Use:   "start",
	Short: "start Andromeda server",
	Run: func(cmd *cobra.Command, args []string) {
		rand.Seed(time.Now().Unix())
		files, err := ioutil.ReadDir(folder)
		if err != nil {
			log.Fatal(err)
		}

		cli := challenge.CreateDockerClient()

		for _, file := range files {
			if !file.IsDir() {
				continue
			}
			_, err := challenge.Create(folder + file.Name())
			if err != nil {
				fmt.Println("Error processing challenge: "+file.Name(), err)
			}
		}

		instances := challenge.StartServer(&cli, bindIp)

		fmt.Println("Listening on", apiAddress)
		server := api.Server{
			Instances: instances,
		}
		(&server).StartServer(apiAddress)
	},
}

var apiAddress string
var bindIp string

func init() {
	StartCommand.Flags().StringVarP(&apiAddress, "api_address", "a", "127.0.0.1:6000", "api_address")
	StartCommand.Flags().StringVarP(&bindIp, "bind_ip", "b", "127.0.0.1", "bind_ip")
	RootCommand.AddCommand(StartCommand)
}
