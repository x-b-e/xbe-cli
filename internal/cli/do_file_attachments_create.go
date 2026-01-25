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

type doFileAttachmentsCreateOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	FileName       string
	ObjectKey      string
	AttachedToType string
	AttachedToID   string
}

func newDoFileAttachmentsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a file attachment",
		Long: `Create a new file attachment.

Required flags:
  --file-name    File name (required)
  --object-key   S3 object key (required)

Optional flags:
  --attached-to-type  Attached resource type (e.g., projects)
  --attached-to-id    Attached resource ID`,
		Example: `  # Create an unattached file attachment
  xbe do file-attachments create --file-name "invoice.pdf" --object-key "uploads/123/invoice.pdf"

  # Create a file attachment linked to a project
  xbe do file-attachments create \
    --file-name "plan.pdf" \
    --object-key "uploads/456/plan.pdf" \
    --attached-to-type projects \
    --attached-to-id 789

  # Get JSON output
  xbe do file-attachments create --file-name "image.png" --object-key "uploads/abc/image.png" --json`,
		Args: cobra.NoArgs,
		RunE: runDoFileAttachmentsCreate,
	}
	initDoFileAttachmentsCreateFlags(cmd)
	return cmd
}

func init() {
	doFileAttachmentsCmd.AddCommand(newDoFileAttachmentsCreateCmd())
}

func initDoFileAttachmentsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("file-name", "", "File name (required)")
	cmd.Flags().String("object-key", "", "S3 object key (required)")
	cmd.Flags().String("attached-to-type", "", "Attached resource type")
	cmd.Flags().String("attached-to-id", "", "Attached resource ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoFileAttachmentsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoFileAttachmentsCreateOptions(cmd)
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

	if opts.FileName == "" {
		err := fmt.Errorf("--file-name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.ObjectKey == "" {
		err := fmt.Errorf("--object-key is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if (opts.AttachedToType != "" && opts.AttachedToID == "") || (opts.AttachedToType == "" && opts.AttachedToID != "") {
		err := fmt.Errorf("--attached-to-type and --attached-to-id must be set together")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"file-name":  opts.FileName,
		"object-key": opts.ObjectKey,
	}

	requestData := map[string]any{
		"type":       "file-attachments",
		"attributes": attributes,
	}

	if opts.AttachedToType != "" && opts.AttachedToID != "" {
		requestData["relationships"] = map[string]any{
			"attached-to": map[string]any{
				"data": map[string]any{
					"type": opts.AttachedToType,
					"id":   opts.AttachedToID,
				},
			},
		}
	}

	requestBody := map[string]any{
		"data": requestData,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/file-attachments", jsonBody)
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

	row := buildFileAttachmentRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created file attachment %s\n", row.ID)
	return nil
}

func parseDoFileAttachmentsCreateOptions(cmd *cobra.Command) (doFileAttachmentsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	fileName, _ := cmd.Flags().GetString("file-name")
	objectKey, _ := cmd.Flags().GetString("object-key")
	attachedToType, _ := cmd.Flags().GetString("attached-to-type")
	attachedToID, _ := cmd.Flags().GetString("attached-to-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doFileAttachmentsCreateOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		FileName:       fileName,
		ObjectKey:      objectKey,
		AttachedToType: attachedToType,
		AttachedToID:   attachedToID,
	}, nil
}

func buildFileAttachmentRowFromSingle(resp jsonAPISingleResponse) fileAttachmentRow {
	attrs := resp.Data.Attributes

	row := fileAttachmentRow{
		ID:          resp.Data.ID,
		FileName:    stringAttr(attrs, "file-name"),
		ObjectKey:   stringAttr(attrs, "object-key"),
		CanOptimize: boolAttr(attrs, "can-optimize"),
	}

	if rel, ok := resp.Data.Relationships["attached-to"]; ok && rel.Data != nil {
		row.AttachedToType = rel.Data.Type
		row.AttachedToID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	return row
}
