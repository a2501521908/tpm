package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List stored entries and their credential types",
	Args:    cobra.NoArgs,
	RunE:    runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, _ []string) error {
	store, err := openStore()
	if err != nil {
		return err
	}
	names, err := store.ListNames()
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	for _, n := range names {
		entry, err := store.GetEntry(n)
		if err != nil {
			// Show the name even if it can't be decrypted/parsed.
			fmt.Fprintln(out, n)
			continue
		}
		fmt.Fprintf(out, "%s [%s]\n", n, strings.Join(entry.Kinds(), ", "))
	}
	return nil
}
