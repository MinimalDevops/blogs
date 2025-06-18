package main

import (
	"context"
	"dagger.io/dagger"
)

func main() {
	ctx := context.Background()
	client, err := dagger.Connect(ctx)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	src := client.Host().Directory(".", dagger.HostDirectoryOpts{
		Exclude: []string{"node_modules"},
	})

	tests := client.Container().
		From("python:3.11").
		WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithExec([]string{"python", "-m", "pip", "install", "pytest"}).
		WithExec([]string{"pytest", "-q"})

	if _, err = tests.Stdout(ctx); err != nil {
		panic(err)
	}
}
