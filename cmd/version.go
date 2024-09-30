/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "display the version of the self extracting upgrade tool",
	Long:  `A self extracting upgrade tool that can be used to upgrade a software package`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Version: 0.0.1")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
