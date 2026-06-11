package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zhangshuaike/tpm/internal/config"
	"github.com/zhangshuaike/tpm/internal/keyring"
	"github.com/zhangshuaike/tpm/internal/storage"
)

var (
	initDir    string
	initImport bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize TPM: set the data directory and provision the master key",
	Long: `Initialize TPM on this machine.

It records the data directory (typically inside a cloud sync folder) in the local
env.toml, creates the encrypted store layout, and provisions the AES master key
in the OS keyring. On a secondary device, pass --import to paste the master key
exported from your primary device.`,
	Example: `  tpm init --dir "~/Library/Mobile Documents/com~apple~CloudDocs/tpm-data"
  tpm init --dir "~/OneDrive/tpm-data" --import`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().StringVar(&initDir, "dir", "", "data directory for the encrypted store (e.g. a cloud sync folder)")
	initCmd.Flags().BoolVar(&initImport, "import", false, "import an existing master key (for a secondary device)")
	_ = initCmd.MarkFlagRequired("dir")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, _ []string) error {
	dataDir, err := config.ExpandPath(initDir)
	if err != nil {
		return err
	}

	// Provision the master key in the keyring.
	exists, err := keyring.Exists()
	if err != nil {
		return err
	}

	switch {
	case initImport:
		encoded, err := promptSecret(cmd, "Paste your Master Key (base64) from your primary device: ")
		if err != nil {
			return err
		}
		if _, err := keyring.Import(encoded); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Imported master key into the OS keyring.")
	case exists:
		fmt.Fprintln(cmd.OutOrStdout(), "Master key already present in the OS keyring; reusing it.")
	default:
		if _, err := keyring.Generate(); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "No master key found in keyring. Generated a new one.")
	}

	// Bind the store and create the directory layout.
	key, err := keyring.Get()
	if err != nil {
		return err
	}
	store := storage.NewLocalStorage(key)
	if err := store.Init(dataDir); err != nil {
		return err
	}

	// Persist the local config.
	cfg := &config.Config{Local: config.Local{DataDir: dataDir}}
	if err := cfg.Save(); err != nil {
		return err
	}

	cfgPath, _ := config.Path()
	fmt.Fprintf(cmd.OutOrStdout(), "Data directory: %s\n", dataDir)
	fmt.Fprintf(cmd.OutOrStdout(), "Local config:   %s\n", cfgPath)
	fmt.Fprintln(cmd.OutOrStdout(), "Initialization successful!")
	return nil
}

// promptSecret reads a single line from stdin. Terminal echo suppression is
// intentionally avoided to keep the tool dependency-free and pipe-friendly.
func promptSecret(cmd *cobra.Command, prompt string) (string, error) {
	fmt.Fprint(cmd.OutOrStdout(), prompt)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return "", fmt.Errorf("read input: %w", err)
	}
	return strings.TrimSpace(line), nil
}
