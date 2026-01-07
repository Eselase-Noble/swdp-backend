package execution

import "context"

func StartContainer(cli *client.Client, name, image, volume string) error {
	cfg := &container.Config{
		image:      image,
		WorkingDir: "/workspace",
		Cmd:        []string{"npm", "run", "dev"},
	}

	host := &container.HostConfig{
		Binds: []string{volume + ":/workspace"},
		Resouces: container.Resources{
			Memory:   1024 * 1024 * 1024,
			NanoCPUs: 1_000_000_000,
		},
	}

	_, err := cli.ContainerCreate(context.Background(), cfg, host, nil, nil, name)
	if err != nil {
		return err
	}

	return cli.ContainerStart(context.Background(), name, container.StartOptions{})
}
