package runner

import (
	"context"
	"fmt"
	"github.com/dotneet/codeapi/storage"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/google/uuid"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	specs "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/docker/docker/pkg/stdcopy"
)

type DockerRunner struct {
	client      *client.Client
	imageBucket storage.ImageBucket
}

func NewRunner(bucket storage.ImageBucket) *DockerRunner {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	return &DockerRunner{
		client:      cli,
		imageBucket: bucket,
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

type RunResult struct {
	RunId       string
	Output      string
	ObjectNames []string
}

func (runner *DockerRunner) appendTimeoutHandler(code string) string {
	timeout := 10
	indentedCode := "    " + strings.ReplaceAll(code, "\n", "\n    ")
	return `import signal

def timeout_handler(signum, frame):
	raise Exception("timeout")

def main():
    signal.signal(signal.SIGALRM, timeout_handler)
    signal.alarm(` + strconv.Itoa(timeout) + ")" + "\n\n" + indentedCode + "\n\n" + "try:\n    main()\nexcept Exception as e:\n    print('Timeout')"
}

func (runner *DockerRunner) Run(image string, input string) (*RunResult, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		fmt.Printf("Error generating UUID: %v", err)
		return nil, err
	}
	runId := uuid.String()

	// Create temporary directory
	tmpDir, err := ioutil.TempDir("", "runner")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	config := &container.Config{
		Image:      image,
		Cmd:        []string{"python", "-"},
		Tty:        false,
		OpenStdin:  true,
		StdinOnce:  true,
		WorkingDir: "/mnt/work",
	}
	hostConfig := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:     mount.TypeBind,
				Source:   tmpDir,
				Target:   "/mnt/work",
				ReadOnly: false,
			},
		},
	}
	var networkConfig *network.NetworkingConfig = nil
	var platform *specs.Platform = nil
	randomString := fmt.Sprintf("%06d", rand.Intn(1000000))
	containerName := "python-" + randomString
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

	// Write input to container's stdin
	code := runner.appendTimeoutHandler(input)
	go func() {
		defer attachResponse.CloseWrite()
		io.WriteString(attachResponse.Conn, code)
	}()

	// Read output from container's stdout
	var output strings.Builder
	_, err = stdcopy.StdCopy(&output, &output, attachResponse.Reader)
	if err != nil {
		return nil, err
	}

	// Read output from container's stdout and write it to the pipe
	resultChannel := make(chan string)
	objectNames := make([]string, 0)
	go func() {
		// Remove the container when done
		runner.client.ContainerRemove(context.Background(), response.ID, types.ContainerRemoveOptions{Force: true})

		// Check for .png files in the temporary directory
		filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			ext := filepath.Ext(path)
			if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".bmp" {
				objectName, err := runner.imageBucket.PutObject(runId, path)
				if err != nil {
					return err
				}
				objectNames = append(objectNames, objectName)
			}
			return nil
		})
		if err != nil {
			fmt.Println("Error:", err)
		}
		resultChannel <- "ok"
	}()

	<-resultChannel
	return &RunResult{
		Output:      output.String(),
		RunId:       runId,
		ObjectNames: objectNames,
	}, nil
}
