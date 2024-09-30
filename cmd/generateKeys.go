/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/ploynomail/SelfExtractingUpgrade/logic"
	"github.com/spf13/cobra"
)

var privateKeyPath string

// generateKeysCmd represents the generateKeys command
var generateKeysCmd = &cobra.Command{
	Use:   "generateKeys",
	Short: "Generate a new public/private key pair",
	Long:  `This command will generate a new public/private key pair,Used to verify the signature of the package`,
	Run: func(cmd *cobra.Command, args []string) {
		gk := logic.NewGenerateKeys()
		err := gk.GenerateKeyPair()
		if err != nil {
			FmtError(err)
			return
		}
		if privateKeyPath != "" {
			err = gk.SavePrivateKey(privateKeyPath)
			if err != nil {
				FmtError(err)
				return
			}
		} else {
			FmtPrivateKey(gk.GetPrivateKey())
		}
	},
}

func init() {
	rootCmd.AddCommand(generateKeysCmd)
	generateKeysCmd.Flags().StringVarP(&privateKeyPath, "privateKeyPath", "p", "", "The path to the private key,if not provided a new key will be generated and output to the console")
}
