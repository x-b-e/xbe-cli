package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type fileAttachmentsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type fileAttachmentDetails struct {
	ID                     string `json:"id"`
	FileName               string `json:"file_name"`
	ObjectKey              string `json:"object_key"`
	CanOptimize            bool   `json:"can_optimize"`
	AttachedToType         string `json:"attached_to_type,omitempty"`
	AttachedToID           string `json:"attached_to_id,omitempty"`
	CreatedByID            string `json:"created_by_id,omitempty"`
	SignedURL              string `json:"signed_url,omitempty"`
	SignedURLHead          string `json:"signed_url_head,omitempty"`
	SignedURLOptimized     string `json:"signed_url_optimized,omitempty"`
	SignedURLOptimizedHead string `json:"signed_url_optimized_head,omitempty"`
}

func newFileAttachmentsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show file attachment details",
		Long: `Show the full details of a file attachment.

Output Fields:
  ID                       File attachment identifier
  File Name                Original file name
  Object Key               S3 object key
  Can Optimize             Whether image optimization is available
  Attached To              Attached resource (type/id)
  Created By               Creator user ID
  Signed URL               Download URL (GET)
  Signed URL Head          Signed URL for HEAD requests
  Signed URL Optimized     Download URL for optimized file (if available)
  Signed URL Optimized Head  Signed URL for HEAD requests (optimized)

Arguments:
  <id>    The file attachment ID (required). You can find IDs using the list command.`,
		Example: `  # Show a file attachment
  xbe view file-attachments show 123

  # Get JSON output
  xbe view file-attachments show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runFileAttachmentsShow,
	}
	initFileAttachmentsShowFlags(cmd)
	return cmd
}

func init() {
	fileAttachmentsCmd.AddCommand(newFileAttachmentsShowCmd())
}

func initFileAttachmentsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runFileAttachmentsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseFileAttachmentsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("file attachment id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[file-attachments]", "file-name,object-key,can-optimize,signed-url,signed-url-head,signed-url-optimized,signed-url-optimized-head,attached-to,created-by")

	body, _, err := client.Get(cmd.Context(), "/v1/file-attachments/"+id, query)
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

	details := buildFileAttachmentDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderFileAttachmentDetails(cmd, details)
}

func parseFileAttachmentsShowOptions(cmd *cobra.Command) (fileAttachmentsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return fileAttachmentsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildFileAttachmentDetails(resp jsonAPISingleResponse) fileAttachmentDetails {
	attrs := resp.Data.Attributes

	details := fileAttachmentDetails{
		ID:                     resp.Data.ID,
		FileName:               stringAttr(attrs, "file-name"),
		ObjectKey:              stringAttr(attrs, "object-key"),
		CanOptimize:            boolAttr(attrs, "can-optimize"),
		SignedURL:              stringAttr(attrs, "signed-url"),
		SignedURLHead:          stringAttr(attrs, "signed-url-head"),
		SignedURLOptimized:     stringAttr(attrs, "signed-url-optimized"),
		SignedURLOptimizedHead: stringAttr(attrs, "signed-url-optimized-head"),
	}

	if rel, ok := resp.Data.Relationships["attached-to"]; ok && rel.Data != nil {
		details.AttachedToType = rel.Data.Type
		details.AttachedToID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}

	return details
}

func renderFileAttachmentDetails(cmd *cobra.Command, details fileAttachmentDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.FileName != "" {
		fmt.Fprintf(out, "File Name: %s\n", details.FileName)
	}
	if details.ObjectKey != "" {
		fmt.Fprintf(out, "Object Key: %s\n", details.ObjectKey)
	}
	fmt.Fprintf(out, "Can Optimize: %s\n", yesNo(details.CanOptimize))

	if details.AttachedToType != "" && details.AttachedToID != "" {
		fmt.Fprintf(out, "Attached To: %s/%s\n", details.AttachedToType, details.AttachedToID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	if details.SignedURL != "" {
		fmt.Fprintf(out, "Signed URL: %s\n", details.SignedURL)
	}
	if details.SignedURLHead != "" {
		fmt.Fprintf(out, "Signed URL Head: %s\n", details.SignedURLHead)
	}
	if details.SignedURLOptimized != "" {
		fmt.Fprintf(out, "Signed URL Optimized: %s\n", details.SignedURLOptimized)
	}
	if details.SignedURLOptimizedHead != "" {
		fmt.Fprintf(out, "Signed URL Optimized Head: %s\n", details.SignedURLOptimizedHead)
	}

	return nil
}
