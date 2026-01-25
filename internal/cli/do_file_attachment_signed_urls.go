package cli

import "github.com/spf13/cobra"

var doFileAttachmentSignedUrlsCmd = &cobra.Command{
	Use:   "file-attachment-signed-urls",
	Short: "Generate signed URLs for file attachments",
	Long: `Generate signed URLs for file attachments.

Use this command to create temporary, secure download URLs for existing
file attachments. You must have access to the file attachment or the
resource it is attached to.

Commands:
  create    Generate a signed URL for a file attachment`,
	Example: `  # Generate a signed URL
  xbe do file-attachment-signed-urls create --file-attachment-id 123

  # JSON output
  xbe do file-attachment-signed-urls create --file-attachment-id 123 --json`,
}

func init() {
	doCmd.AddCommand(doFileAttachmentSignedUrlsCmd)
}
