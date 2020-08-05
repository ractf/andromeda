package app

import "github.com/spf13/cobra"

var RootCommand = &cobra.Command{
	Use: "andromeda",
}

var address string
var apiKey string

func init() {
	RootCommand.PersistentFlags().StringVarP(&address, "address", "a", "127.0.0.1:6000", "ip:port")
	RootCommand.PersistentFlags().StringVarP(&apiKey, "api-key", "k", "", "api key")
}
