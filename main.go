package main

import (
	"context"
	"dagger.io/dagger"
)

func main() {
	ctx := context.Background()
	client, err := dagger.Connect(ctx)
	defer client.Close()

	src := client.Host().Directory(".", dagger.HostDirectoryOpts{
		Exclude: []string{"node_modules"},
	})

	ctr := client.Container().
		From("node:18").
		WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithExec([]string{"npm", "ci"}).
		WithExec([]string{"npx", "eslint", "."})

	_, err = ctr.Stdout(ctx)
	if err != nil {
		panic(err)
	}
}
