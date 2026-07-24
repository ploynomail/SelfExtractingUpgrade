package cmd

import (
	"fmt"

	"github.com/ploynomail/SelfExtractingUpgrade/logic"
	"github.com/spf13/cobra"
)

var outputPath string

var generateKeysCmd = &cobra.Command{
	Use:   "generateKeys",
	Short: "Generate a new public/private key pair",
	Long:  `This command will generate a new public/private key pair, used to verify the signature of the package`,
	RunE: func(cmd *cobra.Command, args []string) error {
		gk := logic.NewGenerateKeys()
		if err := gk.GenerateKeyPair(); err != nil {
			return fmt.Errorf("generate key pair: %w", err)
		}
		if outputPath != "" {
			if err := gk.SavePrivateKey(outputPath); err != nil {
				return fmt.Errorf("save keys: %w", err)
			}
			fmt.Printf("Keys saved to %s.key and %s.pub\n", outputPath, outputPath)
			return nil
		}
		return FmtPrivateKey(gk.GetPrivateKey())
	},
}

func init() {
	rootCmd.AddCommand(generateKeysCmd)
	generateKeysCmd.Flags().StringVarP(&outputPath, "output", "o", "", "output path prefix for generated keys (if not provided, keys are printed to stdout)")
}
