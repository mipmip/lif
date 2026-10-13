package commands

import (
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// legacyDataName matches the data names a help text may still quote for the
// legacy fallback: the .beans store and config file names, the BEANS_PATH env
// var and a beans.<key> config key path. Those belong to the legacy-names
// leaves (D18); every other bean/beans word in a help text is a leftover.
var legacyDataName = regexp.MustCompile(`\.beans(\.yml|\.bak-\S*|/\S*)?|BEANS_[A-Z_]+|\bbeans\.[a-z_]+`)

var oldProductWord = regexp.MustCompile(`(?i)\bbeans?\b`)

// SC-01 / AC-01: the help of the root command and of every subcommand names
// the product as lif and the noun as item. prime and graphql carry their own
// texts (prime leaf) and are not part of this check.
func TestHelpTextsNameLifAndItem(t *testing.T) {
	root := sharedTestRoot(t)

	var walk func(cmd *cobra.Command)
	walk = func(cmd *cobra.Command) {
		if cmd.Name() == "prime" || cmd.Name() == "graphql" {
			return
		}
		help := cmd.Short + "\n" + cmd.Long + "\n" + cmd.UsageString()
		for _, line := range strings.Split(help, "\n") {
			rest := legacyDataName.ReplaceAllString(line, "")
			if oldProductWord.MatchString(rest) {
				t.Errorf("%s help names bean/beans: %q", cmd.CommandPath(), strings.TrimSpace(line))
			}
		}
		for _, sub := range cmd.Commands() {
			walk(sub)
		}
	}
	walk(root)
}
