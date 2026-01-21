package cli

import "github.com/spf13/cobra"

var tractorTrailerCredentialClassificationsCmd = &cobra.Command{
	Use:   "tractor-trailer-credential-classifications",
	Short: "View tractor/trailer credential classifications",
	Long: `View tractor and trailer credential classifications.

These classifications define types of credentials that can be assigned to tractors and trailers.

Commands:
  list  List tractor/trailer credential classifications`,
}

func init() {
	viewCmd.AddCommand(tractorTrailerCredentialClassificationsCmd)
}
