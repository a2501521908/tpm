package cmd

import (
	"github.com/spf13/cobra"

	"github.com/zhangshuaike/tpm/internal/storage"
)

var codeSilent bool

var codeCmd = &cobra.Command{
	Use:     "code <name>",
	Aliases: []string{"tempCode", "tempcode", "temp"},
	Short:   "Print the current dynamic code (TOTP) for an entry",
	Long: `Generate and print the current one-time code (TOTP) for the named entry.

Use --silent for AI/script pipelines: it prints ONLY the code with no newline,
color, or extra text, e.g.

  export MY_TOKEN=$(tpm code github --silent)`,
	Example: `  tpm code github
  tpm code github --silent`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return readAndPrint(cmd, args[0], storage.KindCode, codeSilent)
	},
}

func init() {
	codeCmd.Flags().BoolVarP(&codeSilent, "silent", "s", false, "print only the code (no newline/color), ideal for scripts and AI")
	rootCmd.AddCommand(codeCmd)
}
