package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/parnurzeal/gorequest"
)

type Info struct {
	Version string `json:"version"`
}

func main() {
	docker, _ := NewDockerClient()

	containerID, err := GetCurrentContainerID()
	if err != nil {
		panic(err)
	}

	container, err := docker.GetContainer(containerID)
	if err != nil {
		panic(err)
	}

	if docker.IsTransitionalContainer() {
		parent, err := docker.GetParentContainer()
		if err != nil {
			panic(err)
		}

		if err := docker.StopContainer(parent); err != nil {
			panic(err)
		}

		_, err = docker.UpdateContainerImage(parent, container.info.Config.Image, parent.info.Name, false)
		if err != nil {
			panic(err)
		}

		os.Exit(0)
	}

	for {
		fmt.Println("Checking for updates...")

		info := new(Info)
		_, _, _ = gorequest.New().Get("http://localhost:3000/info").EndStruct(&info)

		parts := strings.SplitN(container.info.Config.Image, ":", 2)
		ok, err := IsUpdateAvailable(parts[1], info.Version)
		if err != nil {
			panic(err)
		}

		if ok {
			fmt.Printf("Updating from %s to %s\n", parts[1], info.Version)

			_, err := docker.UpdateContainerImage(container, fmt.Sprintf("%s:%s", parts[0], info.Version), "", true)
			if err != nil {
				panic(err)
			}
		}

		time.Sleep(time.Second * 30)
	}
}
