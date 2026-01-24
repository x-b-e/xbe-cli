package cli

import "github.com/spf13/cobra"

var doFileAttachmentsCmd = &cobra.Command{
	Use:     "file-attachments",
	Aliases: []string{"file-attachment"},
	Short:   "Manage file attachments",
	Long: `Create, update, and delete file attachments.

File attachments store uploaded files and can be linked to other resources
using the attached-to relationship.`,
}

func init() {
	doCmd.AddCommand(doFileAttachmentsCmd)
}
