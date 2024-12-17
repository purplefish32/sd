/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Info for the current scope",

	Run: func(cmd *cobra.Command, args []string) {
		currentInstance := viper.GetString("current-instance")
		currentDevice := viper.GetString("current-device")
		currentProfile := viper.GetString("current-profile")
		currentPage := viper.GetString("current-page")

		fmt.Print("current-instance: ")
		if currentInstance == "" {
			fmt.Println("None")

		} else {
			fmt.Println(currentInstance)
		}

		fmt.Print("current-device: ")
		if currentDevice == "" {
			fmt.Println("None")

		} else {
			fmt.Println(currentDevice)
		}

		fmt.Print("current-profile: ")
		if currentProfile == "" {
			fmt.Println("None")
		} else {
			fmt.Println(currentProfile)
		}

		fmt.Print("current-page: ")
		if currentPage == "" {
			fmt.Println("None")
		} else {
			fmt.Println(currentPage)
		}
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
