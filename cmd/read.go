package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/zhangshuaike/tpm/internal/provider"
)

// readAndPrint loads the entry, selects the credential of the given kind,
// generates its current value and prints it.
func readAndPrint(cmd *cobra.Command, name, kind string, silent bool) error {
	store, err := openStore()
	if err != nil {
		return err
	}
	entry, err := store.GetEntry(name)
	if err != nil {
		return err
	}
	cred, err := entry.Get(kind)
	if err != nil {
		return fmt.Errorf("entry %q has no %q credential (available: %v)", name, kind, entry.Kinds())
	}
	res, err := provider.Generate(context.Background(), kind, cred)
	if err != nil {
		return err
	}
	printResult(cmd, name, res, silent)
	return nil
}

// printResult writes the value either in human mode or in clean pipeline mode.
func printResult(cmd *cobra.Command, name string, res provider.Result, silent bool) {
	out := cmd.OutOrStdout()
	if silent {
		// No newline, no decoration: AI/pipeline-friendly raw output.
		fmt.Fprint(out, res.Code)
		return
	}
	if res.ExpiresIn > 0 {
		fmt.Fprintf(out, "[TPM] Generating code for %s...\n", name)
		fmt.Fprintf(out, "Code: %s (Expires in %ds)\n", res.Code, res.ExpiresIn)
		return
	}
	fmt.Fprintf(out, "[TPM] %s\n", name)
	fmt.Fprintf(out, "Value: %s\n", res.Code)
}
