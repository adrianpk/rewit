package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <input-file>", os.Args[0])
	}

	inputFile := os.Args[1]
	file, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("Error opening input file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var name, email string
	var repos []string

	for scanner.Scan() {
		line := scanner.Text()
		if name == "" {
			name = line
		} else if email == "" {
			email = line
		} else {
			repos = append(repos, line)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading input file: %v", err)
	}

	if name == "" || email == "" || len(repos) == 0 {
		log.Fatalf("Input file must contain a name, an email, and at least one repository URL")
	}

	for _, repo := range repos {
		cloneAndRewriteHistory(repo, name, email)
	}
}

func cloneAndRewriteHistory(repo, name, email string) {
	repoName := getRepoName(repo)
	if repoName == "" {
		log.Fatalf("Invalid repository URL: %s", repo)
	}

	// Clone
	if err := runCommand("git", "clone", "--bare", repo, repoName+".git"); err != nil {
		log.Fatalf("Failed to clone repository %s: %v", repo, err)
	}

	// Change directory
	if err := os.Chdir(repoName + ".git"); err != nil {
		log.Fatalf("Failed to change directory to %s: %v", repoName+".git", err)
	}
	defer os.Chdir("..")

	// Rewrite
	filterBranchCmd := fmt.Sprintf("git filter-branch --env-filter 'GIT_AUTHOR_NAME=\"%s\" GIT_AUTHOR_EMAIL=\"%s\" GIT_COMMITTER_NAME=\"%s\" GIT_COMMITTER_EMAIL=\"%s\"' --tag-name-filter cat -- --all", name, email, name, email)
	if err := runCommand("sh", "-c", filterBranchCmd); err != nil {
		log.Fatalf("Failed to rewrite history for repository %s: %v", repo, err)
	}

	// Force push
	if err := runCommand("git", "push", "--force", "--tags", "origin", "refs/heads/*"); err != nil {
		log.Fatalf("Failed to force push repository %s: %v", repo, err)
	}
}

func getRepoName(repo string) string {
	parts := strings.Split(repo, "/")
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSuffix(parts[len(parts)-1], ".git")
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
