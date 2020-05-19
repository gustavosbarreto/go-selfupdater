package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/labstack/echo"
)

type Info struct {
	Version string `json:"version"`
}

func main() {
	up, err := NewUpdater()
	if err != nil {
		panic(err)
	}

	currentVersion, err := up.CurrentVersion()
	if err != nil {
		panic(err)
	}

	if err := up.CompleteUpdate(); err != nil {
		panic(err)
	}

	fmt.Printf("Ola! Eu sou um container e minha version eh: %s\n", currentVersion.String())

	e := echo.New()
	e.GET("/info", func(c echo.Context) error {
		version, err := ioutil.ReadFile("/VERSION")
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, Info{Version: string(bytes.TrimSpace(version))})
	})

	go func() {
		if err := e.Start(":3000"); err != nil {
			panic(err)
		}
	}()

	for {
		nextVersion, err := CheckUpdate()
		if err != nil {
			panic(err)
		}

		if nextVersion.GreaterThan(currentVersion) {
			up.ApplyUpdate(nextVersion)
		}

		time.Sleep(time.Second * 10)
	}
}
