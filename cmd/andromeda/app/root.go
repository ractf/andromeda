package app

import "github.com/spf13/cobra"

var RootCommand = &cobra.Command{
	Use:   "andromeda",
	Short: "Challenge Server",
}

var folder string

func init() {
	RootCommand.PersistentFlags().StringVarP(&folder, "folder", "f", "", "challenge folder")
}
