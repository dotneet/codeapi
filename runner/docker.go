package runner

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	specs "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/docker/docker/pkg/stdcopy"
)

type DockerRunner struct {
	client *client.Client
}

func NewRunner() *DockerRunner {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	return &DockerRunner{
		client: cli,
	}
}

func (runner *DockerRunner) List() {
	containers, err := runner.client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		fmt.Printf("%s %s\n", container.ID[:10], container.Image)
	}
}

func (runner *DockerRunner) Run(image string, input string) (*io.PipeReader, error) {
	config := &container.Config{
		Image:     image,
		Cmd:       []string{"python", "-"},
		Tty:       false,
		OpenStdin: true,
		StdinOnce: true,
	}
	hostConfig := &container.HostConfig{}
	var networkConfig *network.NetworkingConfig = nil
	var platform *specs.Platform = nil
	containerName := "hello-world"
	response, err := runner.client.ContainerCreate(
		context.Background(),
		config,
		hostConfig,
		networkConfig,
		platform,
		containerName,
	)

	if err != nil {
		return nil, err
	}

	err = runner.client.ContainerStart(context.Background(), response.ID, types.ContainerStartOptions{})
	if err != nil {
		return nil, err
	}

	// Attach to container to provide input and receive output
	attachOptions := types.ContainerAttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
	}
	attachResponse, err := runner.client.ContainerAttach(context.Background(), response.ID, attachOptions)
	if err != nil {
		return nil, err
	}

	reader, writer := io.Pipe()

	// Write input to container's stdin
	go func() {
		defer attachResponse.CloseWrite()
		io.WriteString(attachResponse.Conn, input)
	}()

	// Read output from container's stdout and write it to the pipe
	go func() {
		defer writer.Close()
		_, err := stdcopy.StdCopy(writer, writer, attachResponse.Reader)
		if err != nil {
			writer.CloseWithError(err)
		}

		// Remove the container when done
		runner.client.ContainerRemove(context.Background(), response.ID, types.ContainerRemoveOptions{Force: true})
	}()

	return reader, nil
}
