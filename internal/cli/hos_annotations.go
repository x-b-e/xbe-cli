package cli

import "github.com/spf13/cobra"

var hosAnnotationsCmd = &cobra.Command{
	Use:     "hos-annotations",
	Aliases: []string{"hos-annotation"},
	Short:   "Browse HOS annotations",
	Long: `Browse HOS annotations.

HOS annotations capture comments and metadata for hours-of-service days
and events.

Commands:
  list  List HOS annotations with filtering and pagination
  show  Show full details of an annotation`,
	Example: `  # List HOS annotations
  xbe view hos-annotations list

  # Filter by HOS day
  xbe view hos-annotations list --hos-day 123

  # Show an annotation
  xbe view hos-annotations show 456`,
}

func init() {
	viewCmd.AddCommand(hosAnnotationsCmd)
}
