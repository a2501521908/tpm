package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/zhangshuaike/tpm/internal/storage"
)

var showCmd = &cobra.Command{
	Use:     "show <name>",
	Aliases: []string{"info"},
	Short:   "Show the credential types and metadata for an entry (no secrets)",
	Long:    "Inspect what an entry contains: which credential types it has, plus TTL and description. Secrets are never printed.",
	Example: `  tpm show tencent`,
	Args:    cobra.ExactArgs(1),
	RunE:    runShow,
}

func init() {
	rootCmd.AddCommand(showCmd)
}

func runShow(cmd *cobra.Command, args []string) error {
	name := args[0]
	store, err := openStore()
	if err != nil {
		return err
	}
	entry, err := store.GetEntry(name)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Name: %s\n", entry.Name)
	kinds := entry.Kinds()
	if len(kinds) == 0 {
		fmt.Fprintln(out, "  (no credentials)")
		return nil
	}
	fmt.Fprintln(out, "Types:")
	for _, k := range kinds {
		cred := entry.Credentials[k]
		desc := ""
		if cred.Meta != nil {
			if d, ok := cred.Meta["description"].(string); ok && d != "" {
				desc = d
			}
		}
		switch {
		case k == storage.KindCode && desc != "":
			fmt.Fprintf(out, "  - %s (ttl=%ds) — %s\n", k, ttlOrDefault(cred.TTL), desc)
		case k == storage.KindCode:
			fmt.Fprintf(out, "  - %s (ttl=%ds)\n", k, ttlOrDefault(cred.TTL))
		case desc != "":
			fmt.Fprintf(out, "  - %s — %s\n", k, desc)
		default:
			fmt.Fprintf(out, "  - %s\n", k)
		}
	}
	return nil
}

func ttlOrDefault(ttl int) int {
	if ttl <= 0 {
		return 30
	}
	return ttl
}
