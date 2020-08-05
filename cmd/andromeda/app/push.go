package app

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
)

var Push = &cobra.Command{
	Use:   "push [file]",
	Short: "Pushes a config file",
	Run: func(cmd *cobra.Command, args []string) {
		for _, file := range args {
			config, err := ioutil.ReadFile(file)
			if err != nil {
				fmt.Println("Error reading job", file)
				fmt.Println(err)
			}

			request, err := http.NewRequest("POST", "http://"+address+"/jobs", bytes.NewBuffer(config))
			if err != nil {
				fmt.Println("Error creating request to endpoint", file)
				fmt.Println(err)
			}

			request.Header.Add("Authorization", apiKey)
			response, err := http.DefaultClient.Do(request)
			if err != nil {
				fmt.Println("Error submitting job to endpoint", file)
				fmt.Println(err)
			}

			if response.StatusCode != 200 {
				body, _ := ioutil.ReadAll(response.Body)
				fmt.Println("Error submitting job", file)
				fmt.Println(string(body))
			}

			response.Body.Close()
		}
	},
}

func init() {
	RootCommand.AddCommand(Push)
}
