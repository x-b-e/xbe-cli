package cli

import "github.com/spf13/cobra"

var publicPraiseReactionsCmd = &cobra.Command{
	Use:     "public-praise-reactions",
	Aliases: []string{"public-praise-reaction"},
	Short:   "View public praise reactions",
	Long: `View public praise reactions.

Public praise reactions capture emoji reactions applied to public praises.

Commands:
  list    List public praise reactions
  show    Show public praise reaction details`,
	Example: `  # List public praise reactions
  xbe view public-praise-reactions list

  # Show a public praise reaction
  xbe view public-praise-reactions show 123

  # Output JSON
  xbe view public-praise-reactions list --json`,
}

func init() {
	viewCmd.AddCommand(publicPraiseReactionsCmd)
}
