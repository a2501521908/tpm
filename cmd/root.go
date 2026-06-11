// Package cmd wires the cobra CLI for TPM and connects the config, keyring,
// storage and provider layers together.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/zhangshuaike/tpm/internal/config"
	"github.com/zhangshuaike/tpm/internal/keyring"
	"github.com/zhangshuaike/tpm/internal/storage"
)

// version is overridable at build time via -ldflags "-X .../cmd.version=...".
var version = "0.1.0"

var bareSilent bool

var rootCmd = &cobra.Command{
	Use:   "tpm [name]",
	Short: "TPM - a terminal & AI-friendly temporary password manager",
	Long: `TPM (Terminal Password Manager) is a lightweight, cross-platform CLI for
temporary/dynamic passwords (TOTP, tokens) designed for terminals and AI agents.

Read a value with an explicit type:   tpm code <name>   /   tpm password <name>
Or, when a name has only one type:    tpm <name>

Encrypted entries are stored as individual files inside a cloud sync folder
(iCloud, OneDrive, Google Drive); the master key never leaves the OS keyring.`,
	Example: `  tpm code github        # explicit type
  tpm github             # shortcut when unambiguous
  tpm add code github --secret "JBSWY3DPEHPK3PXP"`,
	Version:       version,
	SilenceUsage:  true,
	SilenceErrors: true,
	Args:          cobra.ArbitraryArgs,
	RunE:          runBare,
}

func init() {
	rootCmd.Flags().BoolVarP(&bareSilent, "silent", "s", false, "print only the value (no newline/color), ideal for scripts and AI")
}

// runBare handles `tpm <name>`: when the named entry has exactly one credential
// it is printed directly; otherwise the user is asked to specify the type.
func runBare(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}
	if len(args) > 1 {
		return fmt.Errorf("unknown command %q\nusage: tpm <type> <name>  (e.g. tpm code %s)  or  tpm <name>", args[0], args[1])
	}

	name := args[0]
	store, err := openStore()
	if err != nil {
		return err
	}
	entry, err := store.GetEntry(name)
	if err != nil {
		if err == storage.ErrNotFound {
			return fmt.Errorf("no entry named %q (run `tpm list` to see entries, or `tpm --help`)", name)
		}
		return err
	}

	kinds := entry.Kinds()
	switch len(kinds) {
	case 0:
		return fmt.Errorf("entry %q has no credentials", name)
	case 1:
		return readAndPrint(cmd, name, kinds[0], bareSilent)
	default:
		return fmt.Errorf("entry %q has multiple types %v; specify one, e.g. `tpm %s %s`", name, kinds, kinds[0], name)
	}
}

// Execute runs the root command and exits non-zero on error.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

// openStore loads the local config, fetches the master key from the keyring, and
// returns a ready-to-use LocalStorage bound to the configured data dir.
func openStore() (*storage.LocalStorage, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	key, err := keyring.Get()
	if err != nil {
		if err == keyring.ErrNotFound {
			return nil, fmt.Errorf("master key not found in keyring; run `tpm init` first")
		}
		return nil, err
	}
	store := storage.NewLocalStorage(key)
	if err := store.Open(cfg.Local.DataDir); err != nil {
		return nil, err
	}
	return store, nil
}
