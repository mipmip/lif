package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var pathCmd = &cobra.Command{
	Use:   "path",
	Args:  cobra.NoArgs,
	Short: "Print the resolved items directory",
	Long: `Prints the absolute path of the items directory this invocation resolved to.

Resolution follows the usual precedence (--lif-path, then the .lif.yml found
upward or named via --config, then LIF_PATH, then the config directory), so
scripts that need the store can ask for it here instead of re-deriving it and
drifting from the CLI.
A legacy .beans.yml and a legacy BEANS_PATH are read in place of the lif names when those are absent.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(cmd.OutOrStdout(), core.Root())
	},
}

func RegisterPathCmd(root *cobra.Command) {
	root.AddCommand(pathCmd)
}
