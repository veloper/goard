package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "goardctl",
	Short: "Goard CLI — manage projects, issues, and users",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func main() {
	host = os.Getenv("GOARD_HOST")
	pat = os.Getenv("GOARD_PAT")
	if host == "" {
		fmt.Fprintln(os.Stderr, "GOARD_HOST is required (e.g. http://ticketer:8080)")
		os.Exit(1)
	}
	if pat == "" {
		fmt.Fprintln(os.Stderr, "GOARD_PAT is required")
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(usersCmd)
	rootCmd.AddCommand(projectsCmd)
	rootCmd.AddCommand(issuesCmd)
}
