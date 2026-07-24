package cmd

import (
	"fmt"

	"github.com/ploynomail/SelfExtractingUpgrade/logic"
	"github.com/spf13/cobra"
)

var destPath string
var sourcePath string
var senderKey string
var senderPub string
var recipientPub string
var isOverallSign bool

var makeCmd = &cobra.Command{
	Use:   "make",
	Short: "Make a new self extracting upgrade package",
	Long:  `This command will create a new self extracting upgrade package. Encryption and signing are mandatory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if sourcePath == "" {
			return fmt.Errorf("source path (-s) is required")
		}
		if destPath == "" {
			return fmt.Errorf("destination path (-d) is required")
		}
		if senderKey == "" {
			return fmt.Errorf("sender private key (-k) is required")
		}
		if senderPub == "" {
			return fmt.Errorf("sender public key (--sender-pub) is required")
		}
		if recipientPub == "" {
			return fmt.Errorf("recipient public key (-r) is required")
		}

		adca := logic.NewAutoDeCompressAssembly(sourcePath, destPath).
			WithSenderKey(senderKey).
			WithSenderPub(senderPub).
			WithRecipientPub(recipientPub)
		if isOverallSign {
			adca.WithOverallSign()
		}
		return adca.Assembly()
	},
}

func init() {
	rootCmd.AddCommand(makeCmd)
	makeCmd.Flags().StringVarP(&destPath, "dest", "d", "", "destination file name")
	makeCmd.Flags().StringVarP(&sourcePath, "source", "s", "", "source directory path")
	makeCmd.Flags().StringVarP(&senderKey, "sender-key", "k", "", "sender private key file (for signing)")
	makeCmd.Flags().StringVar(&senderPub, "sender-pub", "", "sender public key file (embedded for verification)")
	makeCmd.Flags().StringVarP(&recipientPub, "recipient-pub", "r", "", "recipient public key file (for encrypting session key)")
	makeCmd.Flags().BoolVarP(&isOverallSign, "overall-sign", "O", false, "sign the entire .run file (outputs .sig)")
}
