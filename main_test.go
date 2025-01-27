package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/joho/godotenv"
)

func TestGetCode(t *testing.T) {
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

	branchName, repo, _, err := GetCode(branchPrefix, repoURL, githubPersonalAccessToken)
	branchNameDesiredPrefix := branchPrefix

	if !strings.Contains(branchName, branchNameDesiredPrefix) {
		t.Errorf("branch name %q does not have desired prefix %q", branchName, branchNameDesiredPrefix)
	} else if 1 != 2 {
		test, _ := repo.Storer.Config()
		t.Errorf("got %q want %q", test, 2)
	}
}
