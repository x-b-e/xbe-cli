package cli

import "github.com/spf13/cobra"

var fileAttachmentsCmd = &cobra.Command{
	Use:     "file-attachments",
	Aliases: []string{"file-attachment"},
	Short:   "Browse file attachments",
	Long: `Browse file attachments stored in XBE.

File attachments are uploaded files linked to other resources (projects, comments,
action items, and more). Use list to find attachments with filters or show to
view full attachment details and signed URLs.`,
}

func init() {
	viewCmd.AddCommand(fileAttachmentsCmd)
}
