package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

func main() {
	log.Printf("environment: %v", os.Environ())
	token := getToken()
	owner, repo := getOwnerAndRepo()
	prNumber := getPRNumber()

	ctx := context.Background()
	tokenSrc := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	client := oauth2.NewClient(ctx, tokenSrc)
	if err := run(ctx, client, owner, repo, prNumber, "dependencies", "bors r+"); err != nil {
		workflowFatalLog("Failed to run action: %v", err)
	}
}

func getToken() string {
	token := os.Getenv("INPUT_TOKEN")
	if strings.TrimSpace(token) != "" {
		workflowDebugLog("Found input 'token'. Using that for authentication")
		return token
	}
	workflowDebugLog("Did not find input 'token'. Trying to use GITHUB_TOKEN env var")

	token = os.Getenv("GITHUB_TOKEN")
	if strings.TrimSpace(token) == "" {
		workflowFatalLog("Did not find GITHUB_TOKEN env var. Required for authentication")
	}
	return token
}

func getOwnerAndRepo() (string, string) {
	name := os.Getenv("INPUT_REPOSITORY")
	if strings.TrimSpace(name) == "" {
		workflowFatalLog("repository input required")
	}

	nameParts := strings.Split(name, "/")
	if len(nameParts) != 2 {
		workflowFatalLog("Expected repository name to be of the form <owner>/<repo>, got '%s'", name)
	}
	return nameParts[0], nameParts[1]
}

func getPRNumber() int {
	inputValue := os.Getenv("INPUT_PULL_REQUEST")
	if strings.TrimSpace(inputValue) == "" {
		workflowFatalLog("pull_request input required")
	}

	prNumber, err := strconv.Atoi(inputValue)
	if err != nil {
		workflowFatalLog("pull_request input must be an integer, got %s", inputValue)
	}
	return prNumber
}

func run(ctx context.Context, httpClient *http.Client, owner string, repo string, prNumber int, markerLabel string, comment string) error {
	client := github.NewClient(httpClient)
	pr, _, err := client.Issues.Get(ctx, owner, repo, prNumber)
	if err != nil {
		return fmt.Errorf("failed to get PR %d in %s/%s: %w", prNumber, owner, repo, err)
	}
	isLabeled := false
	for _, label := range pr.Labels {
		if label.GetName() == markerLabel {
			isLabeled = true
			break
		}
	}

	if !isLabeled {
		workflowDebugLog("PR %d in %s/%s is not labeled with marker label '%s'", prNumber, owner, repo, markerLabel)
		return nil
	}

	issueComment := &github.IssueComment{
		Body: &comment,
	}
	_, _, err = client.Issues.CreateComment(ctx, owner, repo, prNumber, issueComment)
	if err != nil {
		return fmt.Errorf("failed to create comment on PR %d in %s/%s: %w", prNumber, owner, repo, err)
	}

	workflowDebugLog("Successfully commented '%s' on PR %d in %s/%s", comment, prNumber, owner, repo)
	return nil
}

// LogLevel represents the level at which to log.
type LogLevel = int

const (
	// DEBUG is the debugging LogLevel.
	DEBUG LogLevel = iota
	// WARNING is the warning LogLevel.
	WARNING
	// ERROR is the error LogLevel
	ERROR
	// FATAL is the fatal LogLevel, like error except exiting afterwards.
	FATAL
)

func workflowLog(level LogLevel, format string, v ...interface{}) {
	var ident string
	switch level {
	case DEBUG:
		ident = "::debug::"
	case WARNING:
		ident = "::warning "
	case ERROR, FATAL:
		ident = "::error "
	default:
		workflowFatalLog("Unknown log level '%d' when attempting to log '%s'", level, format)
	}
	fmt.Printf("%s%s\n", ident, fmt.Sprintf(format, v...))
}

func workflowFatalLog(format string, v ...interface{}) {
	workflowErrorLog(format, v...)
	os.Exit(1)
}

func workflowErrorLog(format string, v ...interface{}) {
	workflowLog(ERROR, format, v...)
}

func workflowWarningLog(format string, v ...interface{}) {
	workflowLog(WARNING, format, v...)
}

func workflowDebugLog(format string, v ...interface{}) {
	workflowLog(DEBUG, format, v...)
}
