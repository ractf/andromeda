package main

import (
	"github.com/ractf/andromeda/cmd/andromeda/app"
)

func main() {
	err := app.RootCommand.Execute()
	if err != nil {
		panic(err)
	}
}
