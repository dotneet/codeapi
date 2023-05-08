package runner

import (
	"context"
	"fmt"
	"github.com/dotneet/codeapi/storage"
	"github.com/samber/lo"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/google/uuid"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type RunResult struct {
	RunId     string
	Output    string
	ImageUrls []string
}

type Runner interface {
	Run(input string) (*RunResult, error)
}

type LanguageSpec interface {
	modifyCodeBeforeRun(code string) string
	createContainer(client *client.Client, tmpDir string) (container.CreateResponse, error)
}

type DockerRunner struct {
	client       *client.Client
	imageBucket  *storage.ImageBucket
	languageSpec LanguageSpec
}

func (runner *DockerRunner) Run(input string) (*RunResult, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		fmt.Printf("Error generating UUID: %v", err)
		return nil, err
	}
	runId := uuid.String()

	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "runner")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	response, err := runner.languageSpec.createContainer(runner.client, tmpDir)
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
	code := runner.languageSpec.modifyCodeBeforeRun(input)
	go func() {
		defer attachResponse.CloseWrite()
		io.WriteString(attachResponse.Conn, code)
	}()

	// Read output from container's stdout
	sbOutput := strings.Builder{}
	_, err = stdcopy.StdCopy(&sbOutput, &sbOutput, attachResponse.Reader)
	if err != nil {
		return nil, err
	}
	output := sbOutput.String()
	inspectResponse, err := runner.client.ContainerInspect(context.Background(), response.ID)
	if err != nil {
		return nil, err
	}
	// Remove the container when done
	runner.client.ContainerRemove(context.Background(), response.ID, types.ContainerRemoveOptions{Force: true})
	if inspectResponse.State.ExitCode != 0 {
		lines := strings.Split(output, "\n")
		return &RunResult{RunId: runId, Output: lines[0]}, nil
	}

	// Read output from container's stdout and write it to the pipe
	imageUrls, err := runner.uploadFiles(runId, tmpDir)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return &RunResult{
		Output:    output,
		RunId:     runId,
		ImageUrls: imageUrls,
	}, nil
}

func (runner *DockerRunner) uploadFiles(runId string, tmpDir string) ([]string, error) {
	// Read output from container's stdout and write it to the pipe
	imageUrls := make([]string, 0)
	extensions := []string{".png", ".jpg", ".jpeg", ".gif", ".bmp", ".csv", ".json", ".txt", ".md", ".mp3", ".wav"}

	// Check for .png files in the temporary directory
	walkErr := filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		ext := filepath.Ext(path)
		if lo.Contains(extensions, ext) {
			key, err := runner.imageBucket.PutObject(runId, path)
			if err != nil {
				return err
			}
			signedUrl, err := runner.imageBucket.GetSignedUrl(key)
			if err != nil {
				return err
			}
			imageUrls = append(imageUrls, signedUrl)
		}
		return nil
	})

	return imageUrls, walkErr
}
