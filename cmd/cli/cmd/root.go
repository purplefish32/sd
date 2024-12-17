/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"sd/cmd/cli/cmd/instance"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sd",
	Short: "sd is a powerful way to control your Stream Deck",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(instance.NewCmd())
}
