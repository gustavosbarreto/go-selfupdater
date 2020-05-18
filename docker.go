package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

type DockerClient interface {
	GetContainer(id string) (*Container, error)
	UpdateContainerImage(container *Container, image, name string, transitional bool) (*Container, error)
	StopContainer(container *Container) error
	IsTransitionalContainer() bool
	GetParentContainer() (*Container, error)
}

type Container struct {
	info *types.ContainerJSON
}

type docker struct {
	api client.CommonAPIClient
}

func NewDockerClient() (DockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	return docker{api: cli}, nil
}

func (d docker) GetContainer(id string) (*Container, error) {
	bg := context.Background()

	info, err := d.api.ContainerInspect(bg, id)
	if err != nil {
		return nil, err
	}

	return &Container{info: &info}, nil
}

func (d docker) UpdateContainerImage(container *Container, image, name string, transitional bool) (*Container, error) {
	ctx := context.Background()

	// Clone container container config and update the image name
	config := container.info.Config
	config.Image = image

	if transitional {
		config.Env = ReplaceOrAppendEnvValues(config.Env, []string{fmt.Sprintf("PARENT_CONTAINER=%s", container.info.ID)})
	}

	netConfig := &network.NetworkingConfig{EndpointsConfig: container.info.NetworkSettings.Networks}

	rd, err := d.api.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return nil, err
	}
	defer rd.Close()

	_, err = io.Copy(ioutil.Discard, rd)
	if err != nil {
		return nil, err
	}

	// Create a new container using the cloned container config
	clone, err := d.api.ContainerCreate(ctx, config, container.info.HostConfig, netConfig, name)
	if err != nil {
		return nil, err
	}

	if err := d.api.ContainerStart(ctx, clone.ID, types.ContainerStartOptions{}); err != nil {
		return nil, err
	}

	return d.GetContainer(clone.ID)
}

func (d docker) StopContainer(container *Container) error {
	ctx := context.Background()

	timeout := time.Minute
	if err := d.api.ContainerStop(ctx, container.info.ID, &timeout); err != nil {
		return err
	}

	opts := types.ContainerRemoveOptions{Force: true, RemoveVolumes: true}
	if err := d.api.ContainerRemove(ctx, container.info.ID, opts); err != nil {
		return err
	}

	return nil
}

func (d docker) IsTransitionalContainer() bool {
	id := os.Getenv("PARENT_CONTAINER")
	return id != ""
}

func (d docker) GetParentContainer() (*Container, error) {
	id := os.Getenv("PARENT_CONTAINER")
	return d.GetContainer(id)
}
