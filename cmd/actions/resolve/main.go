package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	var requested string
	var repo string
	var fallback string

	flag.StringVar(&requested, "requested", "", "requested release tag (use 'latest' to resolve dynamically)")
	flag.StringVar(&repo, "repo", "", "value of github.action_repository")
	flag.StringVar(&fallback, "fallback", "", "value of github.repository (fallback)")
	flag.Parse()

	if repo == "" {
		repo = fallback
	}
	if repo == "" {
		log.Fatal("unable to determine repository that hosts the action")
	}

	version := strings.TrimSpace(requested)
	if version == "" || version == "latest" {
		resolved, err := resolveLatestTag(repo)
		if err != nil {
			log.Fatalf("resolve latest release: %v", err)
		}
		version = resolved
	}

	output := os.Getenv("GITHUB_OUTPUT")
	if output == "" {
		log.Fatal("GITHUB_OUTPUT is not set")
	}

	file, err := os.OpenFile(output, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		log.Fatalf("open GITHUB_OUTPUT: %v", err)
	}
	defer file.Close()

	if _, err := fmt.Fprintf(file, "version=%s\nrepo=%s\n", version, repo); err != nil {
		log.Fatalf("write outputs: %v", err)
	}

	fmt.Printf("Resolved release %q in repository %q\n", version, repo)
}

func resolveLatestTag(repo string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return latestFromList(repo)
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("unexpected HTTP status %s", resp.Status)
	}

	var payload struct {
		Tag string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.Tag == "" {
		return "", fmt.Errorf("latest release response missing tag_name")
	}

	return payload.Tag, nil
}

func latestFromList(repo string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/releases?per_page=1", repo), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("unexpected HTTP status %s", resp.Status)
	}

	var releases []struct {
		Tag   string `json:"tag_name"`
		Draft bool   `json:"draft"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", err
	}
	for _, rel := range releases {
		if rel.Draft {
			continue
		}
		if rel.Tag != "" {
			return rel.Tag, nil
		}
	}
	return "", fmt.Errorf("no published releases found")
}
