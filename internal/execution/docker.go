package execution

import (
	"context"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

func StartContainer(cli *client.Client, name, image, volume string) error {

	ctx := context.Background()

	cfg := &container.Config{
		Image:      image,
		WorkingDir: "/workspace",
		Cmd:        []string{"npm", "run", "dev"},
		Tty:        false,
	}

	host := &container.HostConfig{
		Binds: []string{volume + ":/workspace"},
		Resources: container.Resources{
			Memory:   1024 * 1024 * 1024,
			NanoCPUs: 1_000_000_000,
		},
	}

	//networking := &network.NetworkingConfig{}
	//platform := &v1.Platform{}

	_, err := cli.ContainerCreate(ctx, cfg, host, nil, nil, name)
	if err != nil {
		return err
	}

	return cli.ContainerStart(ctx, name, container.StartOptions{})
}

func StopContainer(cli *client.Client, volume string) {

}
