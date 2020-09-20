package term

import (
	"context"

	"github.com/docker/docker/api/types"
	// "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type (
	DockerClientConn struct {
		types.HijackedResponse
		ctx    context.Context
		client *client.Client
		id     string
	}
	DockerContainerConn struct {
		ID string
	}
)

func NewDockerClientConn(image string) (*DockerClientConn, error) {
	/*	cfg := &container.Config{
			Image: image,
			// Cmd:          sess.Command(),
			// Env:          sess.Environ(),
			Tty:          true,
			OpenStdin:    true,
			AttachStderr: true,
			AttachStdin:  true,
			AttachStdout: true,
			StdinOnce:    true,
			Volumes:      make(map[string]struct{}),
		}
	*/
	docker, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	/*	res, err := docker.ContainerCreate(ctx, cfg, nil, nil, "")
		if err != nil {
			return nil, err
		}*/

	id := "f6bce8924f51"
	opts := types.ContainerAttachOptions{
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Stream: true,
	}
	stream, err := docker.ContainerAttach(ctx, id, opts)
	if err != nil {
		return nil, err
	}

	// err = docker.ContainerStart(ctx, id, types.ContainerStartOptions{})
	return &DockerClientConn{
		HijackedResponse: stream,
		id:               id,
		ctx:              ctx,
		client:           docker,
	}, err
}

func (conn *DockerClientConn) Read(data []byte) (n int, err error) {
	return conn.Reader.Read(data)
}
func (conn *DockerClientConn) Write(data []byte) (int, error) {
	return conn.Conn.Write(data)
}
func (conn *DockerClientConn) Close() error {
	defer conn.client.ContainerRemove(conn.ctx, conn.id, types.ContainerRemoveOptions{})
	return conn.HijackedResponse.CloseWrite()
}

func (conn *DockerClientConn) SendMessage(Message) error {
	return nil
}
func (conn *DockerClientConn) RecoMessage() <- chan Message {
	return nil
}
