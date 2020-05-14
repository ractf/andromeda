package app

import (
	"fmt"
	"github.com/ractf/andromeda/pkg/challenge"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
)

var BuildCommand = &cobra.Command{
	Use:   "build",
	Short: "Builds challenge docker images",
	Run: func(cmd *cobra.Command, args []string) {
		files, err := ioutil.ReadDir(folder)
		if err != nil {
			log.Fatal(err)
		}

		cli := challenge.CreateDockerClient()

		for _, file := range files {
			if !file.IsDir() {
				continue
			}
			spec, err := challenge.Create(folder + file.Name())
			if err != nil {
				fmt.Println("Error processing challenge: "+file.Name(), err)
			}

			err = cli.BuildImage(&spec)
			if err != nil {
				fmt.Println("Error building image: ", err)
			}
		}
	},
}

func init() {
	RootCommand.AddCommand(BuildCommand)
}
