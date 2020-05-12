package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"ractf.co.uk/andromeda/api"
	"ractf.co.uk/andromeda/challenge"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())
	files, err := ioutil.ReadDir(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	cli := challenge.CreateDockerClient()

	for _, file := range files {
		if file.IsDir() {
			_, err := challenge.Create(os.Args[1] + file.Name())
			if err != nil {
				fmt.Println("Error processing challenge: "+file.Name(), err)
			}
			/*
				err = cli.BuildImage(&spec)
				if err != nil {
					fmt.Println("Error building image: ", err)
				}
			*/
		}
	}

	instances := challenge.SetupInstances(&cli)

	fmt.Println("Listening on port 6000")
	server := api.Server{
		Instances: &instances,
	}
	go (&server).StartServer()

	go instances.HousekeepingTick()
	ticker := time.NewTicker(time.Second * time.Duration(30))
	for {
		select {
		case <-ticker.C:
			go instances.HousekeepingTick()
		}
	}
}
