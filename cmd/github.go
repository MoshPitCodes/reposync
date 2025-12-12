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

package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/MoshPitCodes/reposync/internal/github"
	"github.com/MoshPitCodes/reposync/internal/tui"
)

var (
	githubOwner string
	batchMode   bool

	githubCmd = &cobra.Command{
		Use:   "github [repos...]",
		Short: "Synchronize repositories from GitHub",
		Long: `Synchronize repositories from GitHub.
Without arguments, launches an interactive TUI to select repositories.
With --batch flag, clones specified repositories without interaction.`,
		RunE: runGitHub,
	}
)

func init() {
	rootCmd.AddCommand(githubCmd)

	githubCmd.Flags().StringVar(&githubOwner, "owner", "", "GitHub owner/organization (defaults to REPO_SYNC_GITHUB_OWNER env var)")
	githubCmd.Flags().BoolVar(&batchMode, "batch", false, "Batch mode: clone specified repositories without interaction")
}

// runGitHub handles the github subcommand.
func runGitHub(cmd *cobra.Command, args []string) error {
	// Use flag value or fall back to config
	owner := githubOwner
	if owner == "" {
		owner = cfg.GitHubOwner
	}

	if owner == "" {
		return fmt.Errorf("GitHub owner must be specified via --owner flag or REPO_SYNC_GITHUB_OWNER env var")
	}

	// Batch mode: clone specified repos directly
	if batchMode {
		if len(args) == 0 {
			return fmt.Errorf("batch mode requires at least one repository name")
		}

		client, err := github.NewClient()
		if err != nil {
			return fmt.Errorf("failed to initialize GitHub client: %w", err)
		}

		targetDir, err := cfg.GetTargetDir()
		if err != nil {
			return fmt.Errorf("failed to get target directory: %w", err)
		}

		for _, repoName := range args {
			fmt.Printf("Cloning %s/%s...\n", owner, repoName)
			if err := client.CloneRepo(owner, repoName, targetDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error cloning %s: %v\n", repoName, err)
				continue
			}
			fmt.Printf("Successfully cloned %s\n", repoName)
		}

		return nil
	}

	// Interactive mode: launch TUI with GitHub context
	model, err := tui.NewGitHubModel(cfg, owner)
	if err != nil {
		return fmt.Errorf("failed to initialize TUI: %w", err)
	}

	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}
