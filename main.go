package main

import (
	"context"
	"flag"
)

import (
	"github.com/google/go-github/v42/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

func main() {
	// Set log to json format for easier parsing
	log.SetFormatter(&log.JSONFormatter{})

	// Set the CLI flags
	allPtr := flag.Bool("all", true, "Should all personal repos be enabled?")
	orgPtr := flag.String("org", "", "Org to enabled for if 'all' is not set")
	tokenPtr := flag.String("ghToken", "",
		"GitHub token to use to authenticate")
	enablePtr := flag.Bool("enable", true,
		"enable or disable dependabot alerts")
	flag.Parse()

	// Fail without a GitHub token
	if *tokenPtr == "" {
		log.Fatal("GitHub Token is required")
	}

	// We need either all or an org
	if *allPtr == false {
		if *orgPtr == "" {
			log.Fatal("Org is required if 'all' is disabled")
		}
		log.Info("Scanning", *orgPtr)
	}

	// Build a GitHub client
	ctx := context.Background()
	// Authenticate
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *tokenPtr},
	)
	tokenClient := oauth2.NewClient(ctx, tokenSource)
	// Start the client with authentication
	client := github.NewClient(tokenClient)

	// Define any options to use for GitHub
	// We want to paginate, 10 results per page
	optRepos := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 10},
	}
	optUserRepos := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 10},
	}

	// Define `allRepos` so we can use it outside the loops
	var allRepos []*github.Repository
	// Case: Action Specific Org
	if *allPtr == false {
		// list all repositories in given org
		log.Info("Getting Repositories")
		for {
			repos, resp, err := client.Repositories.ListByOrg(ctx, *orgPtr, optRepos)
			if err != nil {
				log.Fatal(err)
			}
			allRepos = append(allRepos, repos...)
			if resp.NextPage == 0 {
				break
			}
			optRepos.Page = resp.NextPage
		}
		log.Debug("Received Repositories:", len(allRepos))
	}

	// Case: Action Personal Repos
	if *allPtr == true {
		// list all repositories for the authenticated user
		log.Info("Getting Repositories")
		for {
			repos, resp, err := client.Repositories.List(ctx, "", optUserRepos)
			if err != nil {
				log.Fatal(err)
			}
			allRepos = append(allRepos, repos...)
			if resp.NextPage == 0 {
				break
			}
			optRepos.Page = resp.NextPage
		}
		log.Debug("Received Repositories:", len(allRepos))
	}

	// Enable Alerts for all Repos
	if *enablePtr == true {
		for _, repo := range allRepos {
			_, err := client.Repositories.EnableVulnerabilityAlerts(ctx, "",
				repo.GetFullName())
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// Disable Alerts for all Repos
	if *enablePtr == false {
		for _, repo := range allRepos {
			_, err := client.Repositories.DisableVulnerabilityAlerts(ctx, "",
				repo.GetFullName())
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
