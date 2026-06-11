package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/zhangshuaike/tpm/internal/config"
	"github.com/zhangshuaike/tpm/internal/provider"
	"github.com/zhangshuaike/tpm/internal/storage"
)

// completeEntryNames suggests existing entry names, filtered by the prefix the
// user has typed so far. It only reads file names from the data dir and never
// touches the keyring, so pressing TAB won't trigger a Keychain prompt.
func completeEntryNames(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	cfg, err := config.Load()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	// A nil key is fine: listing file names does not decrypt anything.
	store := storage.NewLocalStorage(nil)
	if err := store.Open(cfg.Local.DataDir); err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	names, err := store.ListNames()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return filterPrefix(names, toComplete), cobra.ShellCompDirectiveNoFileComp
}

// completeKinds suggests the supported credential types (code, password, ...).
func completeKinds(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return filterPrefix(provider.Kinds(), toComplete), cobra.ShellCompDirectiveNoFileComp
}

func filterPrefix(items []string, prefix string) []string {
	if prefix == "" {
		return items
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		if strings.HasPrefix(it, prefix) {
			out = append(out, it)
		}
	}
	return out
}

// init wires dynamic shell completion onto the commands. It runs after all the
// per-command init() functions have registered the command variables.
func init() {
	// Bare `tpm <name>` completes entry names (alongside subcommand names).
	rootCmd.ValidArgsFunction = completeEntryNames

	// Read commands: first arg is an entry name.
	codeCmd.ValidArgsFunction = completeEntryNames
	passwordCmd.ValidArgsFunction = completeEntryNames
	showCmd.ValidArgsFunction = completeEntryNames
	rmCmd.ValidArgsFunction = completeEntryNames

	// add <type> <name>: type first, then name.
	addCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		switch len(args) {
		case 0:
			return completeKinds(cmd, args, toComplete)
		case 1:
			return completeEntryNames(cmd, nil, toComplete)
		default:
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
	}

	// rename <old> <new>: complete the existing source name only.
	renameCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completeEntryNames(cmd, nil, toComplete)
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// --type flags complete with the supported credential kinds.
	_ = rmCmd.RegisterFlagCompletionFunc("type", completeKinds)
	_ = renameCmd.RegisterFlagCompletionFunc("type", completeKinds)
}
