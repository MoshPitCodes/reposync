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

	"github.com/MoshPitCodes/reposync/internal/local"
	"github.com/MoshPitCodes/reposync/internal/tui"
)

var (
	localCmd = &cobra.Command{
		Use:   "local [paths...]",
		Short: "Synchronize local repositories",
		Long: `Synchronize repositories from local filesystem.
Without arguments, launches an interactive TUI to select repositories from configured source directories.
With --batch flag, copies specified repositories without interaction.`,
		RunE: runLocal,
	}
)

func init() {
	rootCmd.AddCommand(localCmd)

	localCmd.Flags().BoolVar(&batchMode, "batch", false, "Batch mode: copy specified repositories without interaction")
}

// runLocal handles the local subcommand.
func runLocal(cmd *cobra.Command, args []string) error {
	// Batch mode: copy specified repos directly
	if batchMode {
		if len(args) == 0 {
			return fmt.Errorf("batch mode requires at least one repository path")
		}

		scanner := local.NewScanner()
		targetDir, err := cfg.GetTargetDir()
		if err != nil {
			return fmt.Errorf("failed to get target directory: %w", err)
		}

		for _, repoPath := range args {
			fmt.Printf("Copying %s...\n", repoPath)
			if err := scanner.CopyRepo(repoPath, targetDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error copying %s: %v\n", repoPath, err)
				continue
			}
			fmt.Printf("Successfully copied %s\n", repoPath)
		}

		return nil
	}

	// Interactive mode: launch TUI with local context
	model, err := tui.NewLocalModel(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize TUI: %w", err)
	}

	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}
