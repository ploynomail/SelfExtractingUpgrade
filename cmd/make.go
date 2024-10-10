/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/ploynomail/SelfExtractingUpgrade/logic"
	"github.com/spf13/cobra"
)

var destPath string
var sourcePath string
var isSign bool
var OverallSign bool
var isEncrypt bool
var privateKey string
var password string

// makeCmd represents the make command
var makeCmd = &cobra.Command{
	Use:   "make",
	Short: "Make a new self extracting upgrade package",
	Long:  `This command will create a new self extracting upgrade package`,
	Run: func(cmd *cobra.Command, args []string) {
		if sourcePath == "" {
			fmt.Println("source path is required")
		}
		if destPath == "" {
			fmt.Println("destination path is required")
		}
		adca := logic.NewAutoDeCompressAssembly(sourcePath, destPath)
		if isSign {
			if privateKey == "" {
				fmt.Println("private key is required")
			}
			adca.WithSign(privateKey)
		}
		if isEncrypt {
			if password == "" {
				fmt.Println("password is required")
			}
			adca.WithEncrypt(password)
		}
		if OverallSign {
			if privateKey == "" {
				fmt.Println("private key is required")
			}
			adca.WithOverallSign()
		}
		err := adca.Assembly()
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(makeCmd)
	makeCmd.Flags().StringVarP(&destPath, "dest", "d", "", "destination file name")
	makeCmd.Flags().StringVarP(&sourcePath, "source", "s", "", "source file path")
	makeCmd.Flags().BoolVarP(&isSign, "sign", "i", false, "sign the package")
	makeCmd.Flags().BoolVarP(&isEncrypt, "encrypt", "c", false, "encrypt the package")
	makeCmd.Flags().StringVarP(&privateKey, "private-key", "k", "", "private key")
	makeCmd.Flags().StringVarP(&password, "password", "p", "", "password")
	makeCmd.Flags().BoolVarP(&OverallSign, "overall-sign", "o", false, "sign the package with overall sign")
}
