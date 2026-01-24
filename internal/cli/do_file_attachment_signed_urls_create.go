package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doFileAttachmentSignedUrlsCreateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	FileAttachmentID string
}

type fileAttachmentSignedURLRow struct {
	ID               string `json:"id"`
	FileAttachmentID string `json:"file_attachment_id,omitempty"`
	SignedURL        string `json:"signed_url,omitempty"`
}

func newDoFileAttachmentSignedUrlsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Generate a signed URL for a file attachment",
		Long: `Generate a signed URL for a file attachment.

Signed URLs are short-lived, pre-authenticated links for downloading file
attachments. You must have access to the file attachment or the resource it
belongs to.

Required flags:
  --file-attachment-id   File attachment ID`,
		Example: `  # Generate a signed URL
  xbe do file-attachment-signed-urls create --file-attachment-id 123

  # JSON output
  xbe do file-attachment-signed-urls create --file-attachment-id 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoFileAttachmentSignedUrlsCreate,
	}
	initDoFileAttachmentSignedUrlsCreateFlags(cmd)
	return cmd
}

func init() {
	doFileAttachmentSignedUrlsCmd.AddCommand(newDoFileAttachmentSignedUrlsCreateCmd())
}

func initDoFileAttachmentSignedUrlsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("file-attachment-id", "", "File attachment ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("file-attachment-id")
}

func runDoFileAttachmentSignedUrlsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoFileAttachmentSignedUrlsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	attributes := map[string]any{
		"file-attachment-id": opts.FileAttachmentID,
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "file-attachment-signed-urls",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/file-attachment-signed-urls", jsonBody)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	row := buildFileAttachmentSignedURLRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.FileAttachmentID != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "File Attachment: %s\n", row.FileAttachmentID)
	}
	if row.SignedURL != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Signed URL: %s\n", row.SignedURL)
		return nil
	}

	if row.ID != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created file attachment signed url %s\n", row.ID)
		return nil
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Created file attachment signed url")
	return nil
}

func parseDoFileAttachmentSignedUrlsCreateOptions(cmd *cobra.Command) (doFileAttachmentSignedUrlsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	fileAttachmentID, _ := cmd.Flags().GetString("file-attachment-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doFileAttachmentSignedUrlsCreateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		FileAttachmentID: fileAttachmentID,
	}, nil
}

func buildFileAttachmentSignedURLRowFromSingle(resp jsonAPISingleResponse) fileAttachmentSignedURLRow {
	resource := resp.Data
	return fileAttachmentSignedURLRow{
		ID:               resource.ID,
		FileAttachmentID: stringAttr(resource.Attributes, "file-attachment-id"),
		SignedURL:        stringAttr(resource.Attributes, "signed-url"),
	}
}
