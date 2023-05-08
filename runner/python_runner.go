package runner

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/dotneet/codeapi/storage"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"math/rand"
	"strconv"
	"strings"
)

type PythonSpec struct {
}

func NewPythonRunner(bucket storage.ImageBucket) Runner {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	return &DockerRunner{
		client:       cli,
		imageBucket:  &bucket,
		languageSpec: &PythonSpec{},
	}
}

func (runner *PythonSpec) modifyCodeBeforeRun(code string) string {
	timeout := 10
	indentedCode := "    " + strings.ReplaceAll(code, "\n", "\n    ")
	return `import signal
import sys
def timeout_handler(signum, frame):
	raise Exception("Timeout")

def main():
    signal.signal(signal.SIGALRM, timeout_handler)
    signal.alarm(` + strconv.Itoa(timeout) + ")" + "\n\n" + indentedCode + "\n\n" +
		`try:
    main()
except Exception as e:
    print(e)
    raise e
`
}

func (runner *PythonSpec) createContainer(client *client.Client, tmpDir string) (container.CreateResponse, error) {
	image := "python_runner"
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
		Resources: container.Resources{
			Memory:   256 * 1024 * 1024, // 256MB
			NanoCPUs: 1000000000,        // 1 CPU
		},
	}
	var networkConfig *network.NetworkingConfig = nil
	var platform *specs.Platform = nil
	randomString := fmt.Sprintf("%06d", rand.Intn(1000000))
	containerName := "python-" + randomString
	return client.ContainerCreate(
		context.Background(),
		config,
		hostConfig,
		networkConfig,
		platform,
		containerName,
	)
}
