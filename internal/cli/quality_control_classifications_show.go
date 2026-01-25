package cli

import "github.com/spf13/cobra"

func newQualityControlClassificationsShowCmd() *cobra.Command {
	return newGenericShowCmd("quality-control-classifications")
}

func init() {
	qualityControlClassificationsCmd.AddCommand(newQualityControlClassificationsShowCmd())
}
