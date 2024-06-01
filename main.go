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
	genyaml := flag.Bool("genyaml", false, "Generate rewit.yml file")
	do := flag.Bool("do", false, "Do the rewrite")
	inputFile := flag.String("file", "rewit.yml", "Input YAML file containing user info and repository URLs")
	userName := flag.String("name", "", "User name to set in the Git commit history")
	userEmail := flag.String("email", "", "User email to set in the Git commit history")
	include := flag.String("include", "", "Exclude repositories that contain this string")
	exclude := flag.String("exclude", "", "Exclude repositories that contain this string")
	tokenEnvar := flag.String("token-envar", "GITHUB_TOKEN", "Environment variable name containing the GitHub token")

	flag.Parse()

	if (*genyaml && *do) || (!*genyaml && !*do) {
		log.Fatalf("Error: Either genyaml or do flag must be set, but not both")
	}

	token := os.Getenv(*tokenEnvar)
	if token == "" {
		log.Fatalf("Error: No GitHub token found in environment variable %s", *tokenEnvar)
	}

	if *genyaml {
		genYaml(*userName, *userEmail, *include, *exclude, *tokenEnvar)
		return
	}

	if *do {
		processRepos(*inputFile)
	}
}

func genYaml(userName, userEmail, include, exclude, tokenEnvar string) {
	stop := make(chan bool)
	go showProgress(stop)

	defer func() {
		stop <- true
	}()

	repos, err := getRepos(include, exclude, tokenEnvar)
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

// getRepos retrieves a list of repositories for a given user from GitHub.
// It uses the provided user's GitHub token for authentication.
// The function filters the repositories based on the include and exclude parameters.
// If an include string is provided, only repositories whose names contain the include string are returned.
// If an exclude string is provided, any repository whose name contains the exclude string is omitted from the results.
// If both include and exclude strings are provided, the exclude string takes precedence over the include string.
// The function returns a slice of repository names in SSH format and any error encountered during the process.
func getRepos(include, exclude, tokenEnvar string) ([]string, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv(tokenEnvar)},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	opt := &github.RepositoryListOptions{
		Type: "all",
	}

	user := ""
	var allRepos []string
	for {
		// Passing the empty string as user will list repositories for the authenticated user.
		repos, resp, err := client.Repositories.List(ctx, user, opt)
		if err != nil {
			return nil, err
		}

		for _, repo := range repos {
			fullName := repo.GetFullName()
			fmt.Println("Evaluating", fullName)

			shouldInclude := include == "" || strings.Contains(fullName, include)
			shouldExclude := exclude != "" && strings.Contains(fullName, exclude)

			if shouldInclude && !shouldExclude {
				sshRepo := sshURL(fullName)
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

// sshURL constructs the SSH URL for a given GitHub repository.
// It takes a repository path as input and prepends it with the standard GitHub SSH URL prefix.
// The function returns the constructed SSH URL as a string.
func sshURL(repoPath string) string {
	url := "git@github.com:" + repoPath

	if strings.HasSuffix(url, ".git") {
		url = url[:len(url)-4]
	}

	return url
}

// processRepos reads a YAML configuration file, validates the user information and repository list,
// and initiates the process of cloning and rewriting the commit history for each repository.
// The function takes the path to the config file as a parameter.
func processRepos(inputFile string) {
	file, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("Error opening input file: %v", err)
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		log.Fatalf("Error parsing YAML file: %v", err)
	}

	isNameEmpty := config.User.Name == ""
	isEmailEmpty := config.User.Email == ""
	isReposEmpty := len(config.Repos) == 0

	if isNameEmpty || isEmailEmpty || isReposEmpty {
		log.Fatalf("You must provide a name and an email in the input file, and at least one repository URL")
	}

	fmt.Println("This process will clone and rewrite the commit history for the following repositories:")

	for _, repo := range config.Repos {
		fmt.Println(repo)
	}

	if !confirm("Are you sure?") {
		fmt.Println("Operation cancelled.")
		return
	}

	for _, repo := range config.Repos {
		cloneAndRewrite(repo, config.User.Name, config.User.Email)
	}
}

// cloneAndRewrite clones a Git repository, rewrites the commit history to update the author and committer information,
// and force-pushes the changes back to the original repository.
// The function takes the repository URL, user name, and user email as parameters.
func cloneAndRewrite(repo, name, email string) {
	repoName := getRepoName(repo)
	if repoName == "" {
		log.Printf("Invalid repository URL: %s", repo)
		return
	}

	// Clone
	if err := runCommand("git", "clone", "--bare", repo, repoName+".git"); err != nil {
		log.Printf("Failed to clone repository %s: %v", repo, err)
		return
	}

	// Change to repo dir
	if err := os.Chdir(repoName + ".git"); err != nil {
		log.Printf("Failed to change directory to %s: %v", repoName+".git", err)
		return
	}
	defer os.Chdir("..")

	filterBranchCmd := fmt.Sprintf("git filter-branch --env-filter "+
		"'GIT_AUTHOR_NAME=\"%s\" GIT_AUTHOR_EMAIL=\"%s\" GIT_COMMITTER_NAME=\"%s\" GIT_COMMITTER_EMAIL=\"%s\"' "+
		"--tag-name-filter cat -- --all", name, email, name, email)

	if err := runCommand("sh", "-c", filterBranchCmd); err != nil {
		log.Printf("Failed to rewrite history for repository %s: %v", repo, err)
		return
	}

	// Force push the changes
	if err := runCommand("git", "push", "--force", "--tags", "origin", "refs/heads/*"); err != nil {
		log.Printf("Failed to force push repository %s: %v", repo, err)
		return
	}
}

// runCommand executes a system command and returns any error encountered.
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// confirm prompts the user with a yes/no question and returns true if the user responds with 'y' or 'yes',
// and false if the user responds with 'n' or 'no'.
// It takes the prompt message as a parameter.
// Keeps asking until a valid response is given.
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

// getRepoName extracts the repository name from a given repository URL.
// It takes the repository URL as a parameter.
func getRepoName(repo string) string {
	parts := strings.Split(repo, "/")
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSuffix(parts[len(parts)-1], ".git")
}

// showProgress displays a rotating progress indicator in the console.
func showProgress(stop chan bool) {
	for {
		select {
		case <-stop:
			fmt.Print("\r  ")
		default:
			for _, r := range `-\|/` {
				fmt.Printf("\r%c", r)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}
