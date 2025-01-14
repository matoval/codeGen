package codesandbox

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

var ContainerID string
var dockerSocket string

func CreateCodeSandbox(ctx context.Context) *client.Client {
	dockerSocket = "unix:///run/user/1000/docker.sock"

	cli, err := client.NewClientWithOpts(client.WithHost(dockerSocket), client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Printf("ERROR: %v", err.Error())
		return nil
	}
	defer cli.Close()

	reader, err := cli.ImagePull(ctx, "debian:bullseye-slim", image.PullOptions{})
	if err != nil {
		fmt.Printf("ERROR: %v", err.Error())
		return nil
	}
	defer reader.Close()

	io.Copy(os.Stdout, reader)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "debian:bullseye-slim",
		OpenStdin: true,
	}, nil, nil, nil, "")
	if err != nil {
		fmt.Printf("ERROR: %v", err.Error())
		return nil
	}

	ContainerID = resp.ID

	if err := cli.ContainerStart(ctx, ContainerID, container.StartOptions{}); err != nil {
		fmt.Printf("ERROR: %v", err.Error())
		return nil
	}

	setupGoPackage()
	return cli
}

func setupGoPackage() {
	cli, err := client.NewClientWithOpts(client.WithHost(dockerSocket), client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Printf("ERROR: %v", err.Error())
		return
	}

	command := []string{"sh", "-c", "apt-get install -y golang && mkdir codeSandbox && cd codeSandbox && go mod init codeSandbox"}

	execCreateResp, err := cli.ContainerExecCreate(context.Background(), ContainerID, container.ExecOptions{
		Cmd: command,
		Tty: false,
	})
	if err != nil {
		fmt.Printf("ERROR: %v", err.Error())
		return
	}

	err = cli.ContainerExecStart(context.Background(), execCreateResp.ID, container.ExecStartOptions{
		Detach: false,
	})
	if err != nil {
		fmt.Printf("ERROR: %v", err.Error())
		return
	}

	containerJSON, err := cli.ContainerInspect(context.Background(), ContainerID)
	if err != nil {
		fmt.Printf("ERROR: %v", err.Error())
		return
	}

	fmt.Printf("container running: %v\n", containerJSON.State.Running)
	return
}

func RemoveContainer(cli *client.Client, ctx context.Context) {
	<-ctx.Done()

	removeOptions := container.RemoveOptions{
		Force: true,
		RemoveVolumes: true,
	}

	err := cli.ContainerRemove(context.Background(), ContainerID, removeOptions)
	if err != nil {
		fmt.Printf("failed to remove container %v: %v", ContainerID, err)
		return
	}
	fmt.Printf("container %v removed successfully.\n", ContainerID)
}

func AddFileToSandbox(fileName string, content string) {
	fmt.Println(ContainerID)
}
