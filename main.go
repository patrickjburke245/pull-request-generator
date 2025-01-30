package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"math/rand"
	"strconv"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v57/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

func main() {
	// Get values from environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	githubPersonalAccessToken := os.Getenv("GITHUB_TOKEN")
	if githubPersonalAccessToken == "" {
		fmt.Println("Error: GITHUB_TOKEN environment variable not set")
		return
	}

	repoURL := os.Getenv("REPO_URL")
	branchPrefix := os.Getenv("BRANCH_PREFIX")
	authorName := os.Getenv("AUTHOR_NAME")
	authorEmail := os.Getenv("AUTHOR_EMAIL")
	baseBranch := os.Getenv("BASE_BRANCH")
	commitMsg := os.Getenv("COMMIT_MESSAGE")
	prTitle := os.Getenv("PR_TITLE")
	prBody := os.Getenv("PR_BODY")

	bugId := strconv.Itoa(rand.Intn(3000))

	if commitMsg == "<auto>" {
		commitMsg = generateCommit("Generate a short, complete commit message for a Git commit fixing a specific bug with bug ID" + bugId)
	}
	if prTitle == "<auto>" {
		prTitle = generateCommit("Generate a concise pull request title with no placeholders for a GitHub pull request related to the commit" + commitMsg)
	}
	if prBody== "<auto>" {
		prBody = generateCommit("Generate a pull request body with no placeholders for a GitHub pull request with the title " + prTitle)
	}
	if branchPrefix == "<auto>" {
		branchPrefix = generateCommit("Generate a GitHub branch name with no markdown to hold the pull request " + prTitle + " for bug ID " + bugId)
	}


	fmt.Println(commitMsg)
	fmt.Println(prTitle)
	fmt.Println(prBody)
	fmt.Println(branchPrefix)
	
	//Get code
	branchName, repo, auth, err := GetCode(branchPrefix, repoURL, githubPersonalAccessToken)
	if err != nil {
		fmt.Printf("Error retrieving code: %s\n", err)
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
	content, err := os.ReadFile("new_resource.tf")
	err = os.WriteFile(exampleFile, content, 0644)
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
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubPersonalAccessToken})
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

	dirRemovalErr := os.RemoveAll("./terragoat")
	if dirRemovalErr != nil {
		log.Fatal(err)
	}
	fmt.Printf("Successfully removed directory.\n")
}

func GetCode(branchPrefix string, repoURL string, githubPersonalAccessToken string) (string, *git.Repository, *http.BasicAuth, error) {
	// Generate unique branch name
	branchName := fmt.Sprintf("%s-%s", branchPrefix, time.Now().Format("20060102-150405"))
	// branchName := branchPrefix
	// Set up authentication
	auth := &http.BasicAuth{
		Username: "x-access-token",
		Password: githubPersonalAccessToken,
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
		return "", nil, nil, err
	}

	return branchName, repo, auth, nil
}
