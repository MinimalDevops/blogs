package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"dagger.io/dagger"
)

func main() {
	// Parse command line flags
	gitRef := flag.String("ref", "main", "Git reference (branch, tag, or commit hash)")
	repoURL := flag.String("repo", "", "Git repository URL")
	flag.Parse()

	if *repoURL == "" {
		fmt.Println("Error: repository URL is required")
		flag.Usage()
		os.Exit(1)
	}

	ctx := context.Background()
	client, err := dagger.Connect(ctx)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// Clone the repository at the specified reference
	src := client.Git(*repoURL, dagger.GitOpts{
		KeepGitDir: true,
		Reference:  *gitRef,
	}).Branch(*gitRef).Tree()

	tests := client.Container().
		From("python:3.11").
		WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithExec([]string{"python", "-m", "pip", "install", "pytest"}).
		WithExec([]string{"pytest", "--import-mode=importlib", "-q"})

	if _, err = tests.Stdout(ctx); err != nil {
		panic(err)
	}
}
