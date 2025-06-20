package main

import (
	"context"
	"flag"
	"fmt"
	"os/exec"
	"strings"
	"dagger.io/dagger"
)

// getGitRemoteURL returns the remote URL of the current Git repository
func getGitRemoteURL() (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git remote URL: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func main() {
	// Parse command line flags
	gitRef := flag.String("ref", "", "Git reference (branch, tag, commit hash, or PR number like 'pr/123'). If not provided, uses local directory")
	repoURL := flag.String("repo", "", "Git repository URL. If not provided, uses current repository's remote URL when --ref is specified")
	flag.Parse()

	ctx := context.Background()
	client, err := dagger.Connect(ctx)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	var src *dagger.Directory

	// If ref is provided, use Git source
	if *gitRef != "" {
		// Get repository URL - either from flag or current git remote
		url := *repoURL
		if url == "" {
			var err error
			url, err = getGitRemoteURL()
			if err != nil {
				panic(fmt.Errorf("no repository URL provided and failed to get current repository URL: %w", err))
			}
		}

		// Handle PR references (format: pr/123) or specific refs
		ref := *gitRef
		isPR := strings.HasPrefix(ref, "pr/")
		if isPR {
			prNum := strings.TrimPrefix(ref, "pr/")
			ref = fmt.Sprintf("refs/pull/%s/head", prNum)
		}

		// Clone the repository at the specified reference
		container := client.Container().
			From("alpine/git:latest").
			WithMountedCache("/cache", client.CacheVolume("git-cache")).
			WithExec([]string{"git", "clone", url, "/src"})

		if isPR {
			// For PRs, we need to fetch the specific ref
			container = container.
				WithWorkdir("/src").
				WithExec([]string{"git", "fetch", "origin", ref}).
				WithExec([]string{"git", "checkout", "FETCH_HEAD"})
		} else {
			// For branches and tags, we can use checkout directly
			container = container.
				WithWorkdir("/src").
				WithExec([]string{"git", "checkout", ref})
		}

		src = container.Directory("/src")
	} else {
		// Use local directory as source
		src = client.Host().Directory(".", dagger.HostDirectoryOpts{
			Exclude: []string{"node_modules"},
		})
	}

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
