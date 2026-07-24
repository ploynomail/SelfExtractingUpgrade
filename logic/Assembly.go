package logic

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"text/template"

	"github.com/ploynomail/SelfExtractingUpgrade/logic/compress"
	"github.com/ploynomail/SelfExtractingUpgrade/logic/keys"
	signatureverify "github.com/ploynomail/SelfExtractingUpgrade/logic/signatureVerify"
)

type AutoDeCompressAssembly struct {
	TargetPath    string
	Path          string
	SenderKey     string
	SenderPub     string
	RecipientPub  string
	IsOverallSign bool
}

func NewAutoDeCompressAssembly(path, targetpath string) *AutoDeCompressAssembly {
	return &AutoDeCompressAssembly{
		Path:       path,
		TargetPath: targetpath,
	}
}

func (c *AutoDeCompressAssembly) WithSenderKey(senderKey string) *AutoDeCompressAssembly {
	c.SenderKey = senderKey
	return c
}

func (c *AutoDeCompressAssembly) WithSenderPub(senderPub string) *AutoDeCompressAssembly {
	c.SenderPub = senderPub
	return c
}

func (c *AutoDeCompressAssembly) WithRecipientPub(recipientPub string) *AutoDeCompressAssembly {
	c.RecipientPub = recipientPub
	return c
}

func (c *AutoDeCompressAssembly) WithOverallSign() *AutoDeCompressAssembly {
	c.IsOverallSign = true
	return c
}

func (c *AutoDeCompressAssembly) Assembly() error {
	sctm := template.New("AutoDeCompressAssembly")
	sctm, err := sctm.Parse(ScriptTemplate)
	if err != nil {
		return err
	}

	// Load keys
	gk := keys.NewGenerateEcdsaKeys()
	senderPriv, err := gk.LoadPrivateKey(c.SenderKey)
	if err != nil {
		return fmt.Errorf("load sender private key: %w", err)
	}
	senderPub, err := gk.LoadPublicKey(c.SenderPub)
	if err != nil {
		return fmt.Errorf("load sender public key: %w", err)
	}
	if !senderPub.Equal(&senderPriv.PublicKey) {
		return fmt.Errorf("sender public key does not match sender private key")
	}
	recipientPub, err := gk.LoadPublicKey(c.RecipientPub)
	if err != nil {
		return fmt.Errorf("load recipient public key: %w", err)
	}

	// Generate random 32-byte session key for payload AES encryption
	sessionKey := make([]byte, 32)
	if _, err := rand.Read(sessionKey); err != nil {
		return fmt.Errorf("generate session key: %w", err)
	}

	// Compress payload
	comp := compress.NewCompressor(c.Path, c.TargetPath)
	if err := comp.Compress(); err != nil {
		return err
	}

	// Encrypt payload with AES-256-CBC (IV prepended)
	plainPayload, err := os.ReadFile(c.TargetPath)
	if err != nil {
		return err
	}
	cipherPayload, err := encrypt(sessionKey, plainPayload)
	if err != nil {
		return fmt.Errorf("encrypt payload: %w", err)
	}
	if err := os.WriteFile(c.TargetPath, cipherPayload, 0644); err != nil {
		return err
	}

	// Sign cipher payload
	sigFile, err := os.Open(c.TargetPath)
	if err != nil {
		return err
	}
	signature, err := signatureverify.SignFile(senderPriv, sigFile)
	sigFile.Close()
	if err != nil {
		return fmt.Errorf("sign payload: %w", err)
	}

	// Encrypt session key with recipient's public key via ECIES
	ephemeralPubPEM, encryptedKeyHex, ivHex, err := ECIESEncryptSessionKey(recipientPub, sessionKey)
	if err != nil {
		return fmt.Errorf("encrypt session key: %w", err)
	}

	// Read sender public key PEM content to embed
	senderPubPEM, err := os.ReadFile(c.SenderPub)
	if err != nil {
		return fmt.Errorf("read sender public key: %w", err)
	}

	data := struct {
		SenderPub     string
		EphemeralPub  string
		EncryptedKey  string
		IV            string
		Signature     string
	}{
		SenderPub:    string(senderPubPEM),
		EphemeralPub: string(ephemeralPubPEM),
		EncryptedKey: encryptedKeyHex,
		IV:           ivHex,
		Signature:    hex.EncodeToString(signature),
	}

	selfRunfile, err := os.Create(c.TargetPath + ".run")
	if err != nil {
		return fmt.Errorf("create .run file: %w", err)
	}
	if err := selfRunfile.Chmod(0755); err != nil {
		selfRunfile.Close()
		return fmt.Errorf("chmod .run file: %w", err)
	}
	if err := sctm.ExecuteTemplate(selfRunfile, "AutoDeCompressAssembly", data); err != nil {
		selfRunfile.Close()
		return err
	}
	if _, err := selfRunfile.Write(cipherPayload); err != nil {
		selfRunfile.Close()
		return err
	}
	if err := selfRunfile.Sync(); err != nil {
		selfRunfile.Close()
		return err
	}
	selfRunfile.Close()

	if c.IsOverallSign {
		ff, err := os.Open(c.TargetPath + ".run")
		if err != nil {
			return err
		}
		overallSignature, err := signatureverify.SignFile(senderPriv, ff)
		ff.Close()
		if err != nil {
			return err
		}
		sigPath := c.TargetPath + ".run.sig"
		if err := os.WriteFile(sigPath, overallSignature, 0644); err != nil {
			return fmt.Errorf("write overall signature file: %w", err)
		}
		fmt.Printf("Overall signature saved to: %s\n", sigPath)
	}

	return nil
}
