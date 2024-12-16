package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "BodyByRocket CLI",
}

func Execute() error {
	return rootCmd.Execute()
}
