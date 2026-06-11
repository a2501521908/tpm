package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/zhangshuaike/tpm/internal/storage"
)

var (
	renameForce bool
	renameType  string
)

var renameCmd = &cobra.Command{
	Use:     "rename <old> <new>",
	Aliases: []string{"mv"},
	Short:   "Rename an entry, or move a single credential type to another name",
	Long: `Rename an entry, or with --type move just one credential type to another name.

  tpm rename github vk                 # rename the whole entry
  tpm rename tencent tencent-pw --type password   # move only the password to a new name`,
	Example: `  tpm rename github vk
  tpm rename tencent tencent-pw --type password`,
	Args: cobra.ExactArgs(2),
	RunE: runRename,
}

func init() {
	renameCmd.Flags().BoolVar(&renameForce, "force", false, "overwrite the target if it already exists")
	renameCmd.Flags().StringVar(&renameType, "type", "", "move only this credential type to <new> instead of renaming the whole entry")
	rootCmd.AddCommand(renameCmd)
}

func runRename(cmd *cobra.Command, args []string) error {
	oldName, newName := args[0], args[1]

	if err := storage.ValidateName(oldName); err != nil {
		return err
	}
	if err := storage.ValidateName(newName); err != nil {
		return err
	}
	if oldName == newName {
		return fmt.Errorf("old and new name are identical: %q", oldName)
	}

	store, err := openStore()
	if err != nil {
		return err
	}

	if renameType != "" {
		return moveCredential(cmd, store, oldName, newName, renameType)
	}
	return renameEntry(cmd, store, oldName, newName)
}

// renameEntry renames a whole entry (all its credentials).
func renameEntry(cmd *cobra.Command, store *storage.LocalStorage, oldName, newName string) error {
	entry, err := store.GetEntry(oldName)
	if err != nil {
		return err
	}

	if !renameForce {
		exists, err := store.Exists(newName)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("entry %q already exists (use --force to overwrite)", newName)
		}
	}

	entry.Name = newName
	if err := store.SaveEntry(entry); err != nil {
		return err
	}
	if err := store.DeleteEntry(oldName); err != nil {
		_ = store.DeleteEntry(newName) // roll back to avoid a duplicate
		return fmt.Errorf("rename failed while removing old entry: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Renamed entry '%s' -> '%s'.\n", oldName, newName)
	return nil
}

// moveCredential moves one credential type from oldName to newName, creating or
// merging into newName and dropping oldName if it becomes empty.
func moveCredential(cmd *cobra.Command, store *storage.LocalStorage, oldName, newName, kind string) error {
	src, err := store.GetEntry(oldName)
	if err != nil {
		return err
	}
	cred, ok := src.Credentials[kind]
	if !ok {
		return fmt.Errorf("entry %q has no %q credential (available: %v)", oldName, kind, src.Kinds())
	}

	dst, err := store.GetEntry(newName)
	if err != nil {
		if err != storage.ErrNotFound {
			return err
		}
		dst = &storage.Entry{Name: newName}
	}
	if _, exists := dst.Credentials[kind]; exists && !renameForce {
		return fmt.Errorf("entry %q already has a %q credential (use --force to overwrite)", newName, kind)
	}

	dst.Set(kind, cred)
	if err := store.SaveEntry(dst); err != nil {
		return err
	}

	delete(src.Credentials, kind)
	if len(src.Credentials) == 0 {
		if err := store.DeleteEntry(oldName); err != nil {
			return fmt.Errorf("move saved to %q but cleanup of %q failed: %w", newName, oldName, err)
		}
	} else if err := store.SaveEntry(src); err != nil {
		return fmt.Errorf("move saved to %q but updating %q failed: %w", newName, oldName, err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Moved %q credential from '%s' to '%s'.\n", kind, oldName, newName)
	return nil
}
