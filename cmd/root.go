/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "Self Extracting Upgrade",
	Short: "A self extracting upgrade tool",
	Long: `A self extracting upgrade tool that can be used to upgrade a software package,
	You can encrypt and sign the package. 
	The package can be automatically decompressed in the Linux shell, 
	and the install.sh in the package can be executed to achieve the desired effect.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
