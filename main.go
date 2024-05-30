package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

// Config represents the structure of the input YAML file.
type Config struct {
	User  User     `yaml:"user"`
	Repos []string `yaml:"repos"`
}

// User represents the user information.
type User struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
}

func main() {
	inputFile := flag.String("file", "rewit.yml", "Input YAML file containing user info and repository URLs")
	userName := flag.String("name", "", "User name to set in the Git commit history")
	userEmail := flag.String("email", "", "User email to set in the Git commit history")
	user := flag.String("user", "", "GitHub user or organization name")
	filter := flag.String("filter", "", "Filter to apply to repository names")
	genyaml := flag.Bool("genyaml", false, "Generate rewit.yml file")

	flag.Parse()

	if *genyaml {
		genYaml(*user, *filter, *userName, *userEmail)
		return
	}

	file, err := os.Open(*inputFile)
	if err != nil {
		log.Fatalf("Error opening input file: %v", err)
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		log.Fatalf("Error parsing YAML file: %v", err)
	}

	if *userName != "" {
		config.User.Name = *userName
	}

	if *userEmail != "" {
		config.User.Email = *userEmail
	}

	if config.User.Name == "" || config.User.Email == "" || len(config.Repos) == 0 {
		log.Fatalf("You must provide a name and an email either in the input file or via flags, and at least one repository URL")
	}

	fmt.Println("This process will clone and rewrite the commit history for the following repositories:")

	for _, repo := range config.Repos {
		fmt.Println(repo)
	}

	if confirm("Do you want to proceed?") {
		for _, repo := range config.Repos {
			cloneAndRewrite(repo, config.User.Name, config.User.Email)
		}
	} else {
		fmt.Println("Process cancelled by the user.")
	}
}

func cloneAndRewrite(repo, name, email string) {
	repoName := getRepoName(repo)
	if repoName == "" {
		log.Fatalf("Invalid repository URL: %s", repo)
	}

	// Clone
	if err := runCommand("git", "clone", "--bare", repo, repoName+".git"); err != nil {
		log.Fatalf("Failed to clone repository %s: %v", repo, err)
	}

	// Change to repo dir
	if err := os.Chdir(repoName + ".git"); err != nil {
		log.Fatalf("Failed to change directory to %s: %v", repoName+".git", err)
	}
	defer os.Chdir("..")

	// Rewrite commit history
	filterBranchCmd := fmt.Sprintf("git filter-branch --env-filter 'GIT_AUTHOR_NAME=\"%s\" GIT_AUTHOR_EMAIL=\"%s\" GIT_COMMITTER_NAME=\"%s\" GIT_COMMITTER_EMAIL=\"%s\"' --tag-name-filter cat -- --all", name, email, name, email)
	if err := runCommand("sh", "-c", filterBranchCmd); err != nil {
		log.Fatalf("Failed to rewrite history for repository %s: %v", repo, err)
	}

	// Force push the changes
	if err := runCommand("git", "push", "--force", "--tags", "origin", "refs/heads/*"); err != nil {
		log.Fatalf("Failed to force push repository %s: %v", repo, err)
	}
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func confirm(msg string) bool {
	fmt.Printf("%s (y/n): ", msg)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		answer := strings.ToLower(scanner.Text())
		if answer == "y" || answer == "yes" {
			return true
		} else if answer == "n" || answer == "no" {
			return false
		}
		fmt.Printf("Invalid input. Please enter 'y' or 'n': ")
	}
	return false
}

func genYaml(user, filter, userName, userEmail string) {
	stop := make(chan bool)
    go showProgress(stop)

	defer func() {
        stop <- true
    }()
	
	repos, err := getRepos(user, filter)
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Create("rewit.yml")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	if userName == "" {
		userName = "John Doe"
	}
	if userEmail == "" {
		userEmail = "john.doe@mail.com"
	}

	fmt.Printf("Using user name: %s\n", userName)
	fmt.Printf("Using email: %s\n", userEmail)
	fmt.Printf("Processing %d repositories...\n", len(repos))

	config := Config{
		User: User{
			Name:  userName,
			Email: userEmail,
		},
		Repos: repos,
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		log.Fatal(err)
	}

	_, err = file.Write(data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nrewit.yml file has been generated")
}

func getRepos(user, filter string) ([]string, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	opt := &github.RepositoryListOptions{
		Type: "all",
	}

	var allRepos []string
	for {
		repos, resp, err := client.Repositories.List(ctx, user, opt)
		if err != nil {
			return nil, err
		}

		for _, repo := range repos {
			if filter == "" || strings.HasPrefix(repo.GetName(), filter) {
				sshRepo := convertHttpsToSsh(repo.GetFullName())
				allRepos = append(allRepos, sshRepo)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}

func getRepoName(repo string) string {
	parts := strings.Split(repo, "/")
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSuffix(parts[len(parts)-1], ".git")
}

func convertHttpsToSsh(url string) string {
	url = strings.Replace(url, "https://github.com/", "git@github.com:", 1)

	if strings.HasSuffix(url, ".git") {
		url = url[:len(url)-4]
	}

	return url
}

func showProgress(stop chan bool) {
    for {
        select {
        case <-stop:
            return
        default:
            for _, r := range `-\|/` {
                fmt.Printf("\r%c", r)
                time.Sleep(100 * time.Millisecond)
            }
        }
    }
}