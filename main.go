package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

func main() {
	// Get token from environment variable
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		fmt.Println("Error: GITHUB_TOKEN environment variable not set")
		return
	}

	// Configuration
	const (
		repoURL      = "https://github.com/patrickjburke245/terragoat.git"
		branchPrefix = "feature"
		authorName   = "Patrick Burke"
		authorEmail  = "24pburke@gmail.com"
		baseBranch   = "master" // Changed from "main" to "master"
		commitMsg    = "Add example changes"
		prTitle      = "Feature: Add example changes"
		prBody       = "This PR adds example changes to the repository."
	)

	// Generate unique branch name
	branchName := fmt.Sprintf("%s-%s", branchPrefix, time.Now().Format("20060102-150405"))

	// Set up authentication
	auth := &http.BasicAuth{
		Username: "x-access-token",
		Password: githubToken,
	}

	// Clone repository
	fmt.Println("Cloning repository...")
	repo, err := git.PlainClone("./terragoat", false, &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
		Auth:     auth,
	})
	if err != nil && err != git.ErrRepositoryAlreadyExists {
		fmt.Printf("Error cloning: %s\n", err)
		return
	}

	// Open repository if it already exists
	if err == git.ErrRepositoryAlreadyExists {
		repo, err = git.PlainOpen("./terragoat")
		if err != nil {
			fmt.Printf("Error opening repository: %s\n", err)
			return
		}
	}

	// Create and checkout new branch
	w, err := repo.Worktree()
	if err != nil {
		fmt.Printf("Error getting worktree: %s\n", err)
		return
	}

	fmt.Printf("Creating and checking out branch: %s\n", branchName)
	b := plumbing.NewBranchReferenceName(branchName)
	err = w.Checkout(&git.CheckoutOptions{
		Create: true,
		Branch: b,
		Force:  false,
	})
	if err != nil {
		fmt.Printf("Error creating branch: %s\n", err)
		return
	}

	// Create an example file
	exampleFile := "./terragoat/example2.tf"
	fmt.Printf("Creating file: %s\n", exampleFile)
	err = os.WriteFile(exampleFile, []byte(`
resource "aws_rds_cluster" "app1-rds-cluster" {
    cluster_identifier      = "app1-rds-cluster"
    allocated_storage       = 10
    backup_retention_period = 0
    storage_encrypted       = false
    
    tags = {
        environment = "development"
        managed_by  = "terraform"
    }
}
`), 0644)
	if err != nil {
		fmt.Printf("Error creating file: %s\n", err)
		return
	}

	// Stage changes
	fmt.Println("Staging changes...")
	_, err = w.Add(".")
	if err != nil {
		fmt.Printf("Error staging files: %s\n", err)
		return
	}

	// Create commit
	fmt.Println("Creating commit...")
	_, err = w.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  authorName,
			Email: authorEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		fmt.Printf("Error committing: %s\n", err)
		return
	}

	// Push changes
	fmt.Println("Pushing changes...")
	err = repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", branchName, branchName)),
		},
		Auth: auth,
	})
	if err != nil {
		fmt.Printf("Error pushing: %s\n", err)
		return
	}

	// Create PR
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubToken})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Extract owner and repo from URL
	parts := strings.Split(strings.TrimSuffix(repoURL, ".git"), "/")
	owner := parts[len(parts)-2]
	repoName := parts[len(parts)-1]

	newPR := &github.NewPullRequest{
		Title: github.String(prTitle),
		Head:  github.String(branchName),
		Base:  github.String(baseBranch),
		Body:  github.String(prBody),
	}

	pr, _, err := client.PullRequests.Create(ctx, owner, repoName, newPR)
	if err != nil {
		fmt.Printf("Error creating PR: %s\n", err)
		return
	}

	fmt.Printf("Successfully created PR #%d\n", pr.GetNumber())
	fmt.Printf("PR URL: %s\n", pr.GetHTMLURL())
}
