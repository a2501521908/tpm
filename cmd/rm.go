package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rmType string

var rmCmd = &cobra.Command{
	Use:     "rm <name>",
	Aliases: []string{"remove", "delete"},
	Short:   "Delete an entry, or a single credential type with --type",
	Long: `Delete a stored entry entirely, or remove just one credential type from it.

  tpm rm github              # delete the whole entry
  tpm rm tencent --type password   # remove only the password, keep the code`,
	Args: cobra.ExactArgs(1),
	RunE: runRm,
}

func init() {
	rmCmd.Flags().StringVar(&rmType, "type", "", "remove only this credential type (e.g. code, password)")
	rootCmd.AddCommand(rmCmd)
}

func runRm(cmd *cobra.Command, args []string) error {
	name := args[0]
	store, err := openStore()
	if err != nil {
		return err
	}

	// Whole-entry deletion.
	if rmType == "" {
		if err := store.DeleteEntry(name); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Deleted entry '%s'.\n", name)
		return nil
	}

	// Single-credential deletion.
	entry, err := store.GetEntry(name)
	if err != nil {
		return err
	}
	if _, ok := entry.Credentials[rmType]; !ok {
		return fmt.Errorf("entry %q has no %q credential (available: %v)", name, rmType, entry.Kinds())
	}
	delete(entry.Credentials, rmType)

	if len(entry.Credentials) == 0 {
		// No credentials left: drop the whole entry.
		if err := store.DeleteEntry(name); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Removed %q (last credential); deleted entry '%s'.\n", rmType, name)
		return nil
	}
	if err := store.SaveEntry(entry); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Removed %q credential from entry '%s'.\n", rmType, name)
	return nil
}
