package cli

import "github.com/spf13/cobra"

var doPublicPraiseReactionsCmd = &cobra.Command{
	Use:     "public-praise-reactions",
	Aliases: []string{"public-praise-reaction"},
	Short:   "Manage public praise reactions",
	Long: `Manage public praise reactions.

Public praise reactions capture emoji reactions applied to public praises.

Commands:
  create    Create a public praise reaction
  delete    Delete a public praise reaction`,
}

func init() {
	doCmd.AddCommand(doPublicPraiseReactionsCmd)
}
