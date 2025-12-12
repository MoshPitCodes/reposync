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

	"github.com/MoshPitCodes/reposync/internal/config"
	"github.com/MoshPitCodes/reposync/internal/tui"
)

var (
	cfg *config.Config

	rootCmd = &cobra.Command{
		Use:   "repo-sync",
		Short: "Repository synchronization tool with interactive TUI",
		Long: `repo-sync is a CLI tool for synchronizing repositories from GitHub or local sources.
It provides an interactive terminal UI powered by Bubble Tea for easy repository management.`,
		RunE: runInteractive,
	}
)

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig loads configuration from environment variables.
func initConfig() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}
}

// runInteractive launches the interactive TUI menu.
func runInteractive(cmd *cobra.Command, args []string) error {
	model, err := tui.NewModel(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize TUI: %w", err)
	}

	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}
