// Copyright 2024-2025 MoshPitCodes
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package github

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
)

// Repository represents a GitHub repository with relevant metadata.
type Repository struct {
	Name          string
	FullName      string
	Description   string
	Language      string
	Stars         int
	CloneURL      string
	IsPrivate     bool
	IsArchived    bool
	DefaultBranch string
}

// TreeEntry represents a single entry in a repository tree.
type TreeEntry struct {
	Path string `json:"path"`
	Type string `json:"type"` // "blob" for files, "tree" for directories
	SHA  string `json:"sha"`
	Size int64  `json:"size,omitempty"`
}

// TreeResponse represents the response from GitHub's Git Trees API.
type TreeResponse struct {
	SHA       string      `json:"sha"`
	Truncated bool        `json:"truncated"`
	Entries   []TreeEntry `json:"tree"`
}

// Client handles GitHub API interactions using go-gh.
type Client struct {
	client *api.RESTClient
}

// NewClient creates a new GitHub client using the existing gh CLI authentication.
func NewClient() (*Client, error) {
	opts := api.ClientOptions{}
	client, err := api.NewRESTClient(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub REST client: %w", err)
	}

	return &Client{client: client}, nil
}

// ListUserRepos retrieves all repositories for a user.
func (c *Client) ListUserRepos(username string) ([]Repository, error) {
	return c.listRepos(fmt.Sprintf("users/%s/repos", username))
}

// ListOrgRepos retrieves all repositories for an organization.
func (c *Client) ListOrgRepos(orgName string) ([]Repository, error) {
	return c.listRepos(fmt.Sprintf("orgs/%s/repos", orgName))
}

// listRepos is a generic method to list repositories from an API endpoint.
func (c *Client) listRepos(endpoint string) ([]Repository, error) {
	var allRepos []Repository
	page := 1
	perPage := 100

	for {
		var repos []struct {
			Name        string `json:"name"`
			FullName    string `json:"full_name"`
			Description string `json:"description"`
			Language    string `json:"language"`
			Stars       int    `json:"stargazers_count"`
			CloneURL    string `json:"clone_url"`
			SSHURL      string `json:"ssh_url"`
			Private     bool   `json:"private"`
			Archived    bool   `json:"archived"`
		}

		url := fmt.Sprintf("%s?per_page=%d&page=%d&sort=updated&direction=desc", endpoint, perPage, page)

		err := c.client.Get(url, &repos)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch repositories: %w", err)
		}

		if len(repos) == 0 {
			break
		}

		for _, repo := range repos {
			allRepos = append(allRepos, Repository{
				Name:        repo.Name,
				FullName:    repo.FullName,
				Description: repo.Description,
				Language:    repo.Language,
				Stars:       repo.Stars,
				CloneURL:    repo.SSHURL, // Prefer SSH for authenticated access
				IsPrivate:   repo.Private,
				IsArchived:  repo.Archived,
			})
		}

		page++
	}

	return allRepos, nil
}

// CloneRepo clones a repository to the specified target directory.
func (c *Client) CloneRepo(owner, repoName, targetDir string) error {
	// Construct the SSH clone URL
	cloneURL := fmt.Sprintf("git@github.com:%s/%s.git", owner, repoName)

	repoPath := filepath.Join(targetDir, repoName)

	// Check if directory already exists
	if _, err := os.Stat(repoPath); err == nil {
		return fmt.Errorf("repository directory already exists: %s", repoPath)
	}

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Clone the repository using git command
	cmd := exec.Command("git", "clone", cloneURL, repoPath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Include git's error output in the error message
		errMsg := strings.TrimSpace(string(output))
		if errMsg != "" {
			return fmt.Errorf("git clone failed: %s", errMsg)
		}
		return fmt.Errorf("git clone failed: %w", err)
	}

	return nil
}

// CloneRepos clones multiple repositories concurrently with progress reporting.
func (c *Client) CloneRepos(repos []Repository, targetDir string, progressFn func(repo string, success bool, err error)) {
	for _, repo := range repos {
		owner := strings.Split(repo.FullName, "/")[0]
		err := c.CloneRepo(owner, repo.Name, targetDir)
		progressFn(repo.Name, err == nil, err)
	}
}

// GetRepoDetails fetches detailed information about a specific repository.
func (c *Client) GetRepoDetails(owner, repoName string) (*Repository, error) {
	var repo struct {
		Name        string `json:"name"`
		FullName    string `json:"full_name"`
		Description string `json:"description"`
		Language    string `json:"language"`
		Stars       int    `json:"stargazers_count"`
		CloneURL    string `json:"clone_url"`
		SSHURL      string `json:"ssh_url"`
		Private     bool   `json:"private"`
		Archived    bool   `json:"archived"`
	}

	endpoint := fmt.Sprintf("repos/%s/%s", owner, repoName)

	if err := c.client.Get(endpoint, &repo); err != nil {
		return nil, fmt.Errorf("failed to fetch repository details: %w", err)
	}

	return &Repository{
		Name:        repo.Name,
		FullName:    repo.FullName,
		Description: repo.Description,
		Language:    repo.Language,
		Stars:       repo.Stars,
		CloneURL:    repo.SSHURL,
		IsPrivate:   repo.Private,
		IsArchived:  repo.Archived,
	}, nil
}

// IsAuthenticated checks if the user is authenticated with GitHub CLI.
func IsAuthenticated() bool {
	cmd := exec.Command("gh", "auth", "status")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// GetCurrentUser retrieves the authenticated user's login.
func (c *Client) GetCurrentUser() (string, error) {
	var user struct {
		Login string `json:"login"`
	}

	if err := c.client.Get("user", &user); err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}

	return user.Login, nil
}

// SearchRepos searches for repositories matching a query.
func (c *Client) SearchRepos(query string, owner string) ([]Repository, error) {
	var result struct {
		Items []struct {
			Name        string `json:"name"`
			FullName    string `json:"full_name"`
			Description string `json:"description"`
			Language    string `json:"language"`
			Stars       int    `json:"stargazers_count"`
			CloneURL    string `json:"clone_url"`
			SSHURL      string `json:"ssh_url"`
			Private     bool   `json:"private"`
			Archived    bool   `json:"archived"`
		} `json:"items"`
	}

	searchQuery := fmt.Sprintf("user:%s %s", owner, query)
	endpoint := fmt.Sprintf("search/repositories?q=%s&sort=updated&order=desc", searchQuery)

	response := make([]byte, 0)
	err := c.client.Get(endpoint, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to search repositories: %w", err)
	}

	if err := json.Unmarshal(response, &result); err != nil {
		return nil, fmt.Errorf("failed to parse search results: %w", err)
	}

	var repos []Repository
	for _, item := range result.Items {
		repos = append(repos, Repository{
			Name:        item.Name,
			FullName:    item.FullName,
			Description: item.Description,
			Language:    item.Language,
			Stars:       item.Stars,
			CloneURL:    item.SSHURL,
			IsPrivate:   item.Private,
			IsArchived:  item.Archived,
		})
	}

	return repos, nil
}

// ListUserOrgs retrieves all organizations for the authenticated user.
func (c *Client) ListUserOrgs() ([]string, error) {
	var orgs []struct {
		Login string `json:"login"`
	}

	err := c.client.Get("user/orgs", &orgs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user organizations: %w", err)
	}

	var orgNames []string
	for _, org := range orgs {
		orgNames = append(orgNames, org.Login)
	}

	return orgNames, nil
}

// RefreshRepo performs a git pull on an existing repository.
func (c *Client) RefreshRepo(repoPath string) error {
	// Verify the directory exists and is a git repository
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("not a git repository: %s", repoPath)
	}

	// Run git pull
	cmd := exec.Command("git", "-C", repoPath, "pull")
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Include git's error output in the error message
		errMsg := strings.TrimSpace(string(output))
		if errMsg != "" {
			return fmt.Errorf("git pull failed: %s", errMsg)
		}
		return fmt.Errorf("git pull failed: %w", err)
	}

	return nil
}

// GetDefaultBranch returns the default branch of a repository.
func (c *Client) GetDefaultBranch(owner, repo string) (string, error) {
	var result struct {
		DefaultBranch string `json:"default_branch"`
	}

	endpoint := fmt.Sprintf("repos/%s/%s", owner, repo)

	if err := c.client.Get(endpoint, &result); err != nil {
		return "", fmt.Errorf("failed to fetch repository: %w", err)
	}

	return result.DefaultBranch, nil
}

// GetRepoTree fetches the complete file tree of a repository recursively.
func (c *Client) GetRepoTree(owner, repo, branch string) (*TreeResponse, error) {
	var result TreeResponse

	endpoint := fmt.Sprintf("repos/%s/%s/git/trees/%s?recursive=1", owner, repo, branch)

	if err := c.client.Get(endpoint, &result); err != nil {
		return nil, fmt.Errorf("failed to fetch repository tree: %w", err)
	}

	return &result, nil
}

// GetFileContent fetches the content of a single file from a repository.
// The content is automatically base64 decoded.
func (c *Client) GetFileContent(owner, repo, path, ref string) ([]byte, error) {
	var result struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
		SHA      string `json:"sha"`
		Size     int64  `json:"size"`
	}

	endpoint := fmt.Sprintf("repos/%s/%s/contents/%s?ref=%s", owner, repo, path, ref)

	if err := c.client.Get(endpoint, &result); err != nil {
		return nil, fmt.Errorf("failed to fetch file content: %w", err)
	}

	// GitHub returns content with newlines that need to be stripped
	cleanContent := strings.ReplaceAll(result.Content, "\n", "")

	if result.Encoding == "base64" {
		decoded, err := base64.StdEncoding.DecodeString(cleanContent)
		if err != nil {
			return nil, fmt.Errorf("failed to decode file content: %w", err)
		}
		return decoded, nil
	}

	return []byte(result.Content), nil
}
