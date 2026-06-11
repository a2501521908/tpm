package cmd

import (
	"github.com/spf13/cobra"

	"github.com/zhangshuaike/tpm/internal/storage"
)

var passwordSilent bool

var passwordCmd = &cobra.Command{
	Use:     "password <name>",
	Aliases: []string{"pass", "pw"},
	Short:   "Print the stored static password for an entry",
	Long: `Print the static password/token stored for the named entry.

Use --silent for AI/script pipelines to emit only the value with no newline.`,
	Example: `  tpm password github
  tpm password github --silent`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return readAndPrint(cmd, args[0], storage.KindPassword, passwordSilent)
	},
}

func init() {
	passwordCmd.Flags().BoolVarP(&passwordSilent, "silent", "s", false, "print only the value (no newline/color)")
	rootCmd.AddCommand(passwordCmd)
}
