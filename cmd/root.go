package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "seu",
	Short: "A self extracting upgrade tool",
	Long: `A self extracting upgrade tool that can be used to upgrade a software package.
You can encrypt and sign the package. 
The package can be automatically decompressed in the Linux shell, 
and the install.sh in the package can be executed to achieve the desired effect.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
