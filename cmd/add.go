package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zhangshuaike/tpm/internal/provider"
	"github.com/zhangshuaike/tpm/internal/storage"
)

var (
	addSecret string
	addTTL    int
	addDesc   string
	addForce  bool
)

var addCmd = &cobra.Command{
	Use:   "add <type> <name>",
	Short: "Add (or update) a credential of a given type on an entry",
	Long: `Add a credential to an entry, encrypted at rest.

The type is the credential kind: "code" (dynamic TOTP) or "password" (static).
A single name can hold several types, so you can run both:
  tpm add code     tencent --secret "JBSWY3DPEHPK3PXP"
  tpm add password tencent --secret "s3cr3t!"`,
	Example: `  tpm add code github --secret "JBSWY3DPEHPK3PXP" --desc "GitHub 2FA"
  tpm add password tencent --secret "s3cr3t!"`,
	Args: cobra.ExactArgs(2),
	RunE: runAdd,
}

func init() {
	addCmd.Flags().StringVar(&addSecret, "secret", "", "the secret value (TOTP base32 seed for 'code', the password itself for 'password')")
	addCmd.Flags().IntVar(&addTTL, "ttl", 30, "TOTP period in seconds (only used by type 'code')")
	addCmd.Flags().StringVar(&addDesc, "desc", "", "optional human-readable description")
	addCmd.Flags().BoolVar(&addForce, "force", false, "overwrite this type if it already exists on the entry")
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	kind, name := args[0], args[1]

	if err := storage.ValidateName(name); err != nil {
		return err
	}
	if !provider.Supported(kind) {
		return fmt.Errorf("unsupported type %q (supported: %v)", kind, provider.Kinds())
	}
	if strings.TrimSpace(addSecret) == "" {
		return fmt.Errorf("--secret is required")
	}

	store, err := openStore()
	if err != nil {
		return err
	}

	// Load existing entry so a new credential merges in; otherwise start fresh.
	entry, err := store.GetEntry(name)
	if err != nil {
		if err != storage.ErrNotFound {
			return err
		}
		entry = &storage.Entry{Name: name}
	}

	if _, exists := entry.Credentials[kind]; exists && !addForce {
		return fmt.Errorf("entry %q already has a %q credential (use --force to overwrite)", name, kind)
	}

	cred := &storage.Credential{Secret: strings.TrimSpace(addSecret)}
	if kind == storage.KindCode {
		cred.TTL = addTTL
	}
	if addDesc != "" {
		cred.Meta = map[string]interface{}{"description": addDesc}
	}
	entry.Set(kind, cred)

	if err := store.SaveEntry(entry); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Saved %q credential on entry '%s'. Encrypted to sync directory.\n", kind, name)
	return nil
}
