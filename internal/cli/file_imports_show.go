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

type fileImportsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type fileImportDetails struct {
	ID               string `json:"id"`
	BrokerID         string `json:"broker_id,omitempty"`
	CreatedByID      string `json:"created_by_id,omitempty"`
	FileAttachmentID string `json:"file_attachment_id,omitempty"`
	ProcessedAt      string `json:"processed_at,omitempty"`
	Note             string `json:"note,omitempty"`
	RawData          any    `json:"raw_data,omitempty"`
}

func newFileImportsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show file import details",
		Long: `Show the full details of a file import.

Output Fields:
  ID
  Broker
  Created By
  File Attachment
  Processed At
  Note
  Raw Data

Arguments:
  <id>    The file import ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a file import
  xbe view file-imports show 123

  # JSON output
  xbe view file-imports show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runFileImportsShow,
	}
	initFileImportsShowFlags(cmd)
	return cmd
}

func init() {
	fileImportsCmd.AddCommand(newFileImportsShowCmd())
}

func initFileImportsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runFileImportsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseFileImportsShowOptions(cmd)
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
		return fmt.Errorf("file import id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[file-imports]", "processed-at,note,raw-data,broker,created-by,file-attachment")
	query.Set("include", "broker,created-by,file-attachment")

	body, _, err := client.Get(cmd.Context(), "/v1/file-imports/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildFileImportDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderFileImportDetails(cmd, details)
}

func parseFileImportsShowOptions(cmd *cobra.Command) (fileImportsShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return fileImportsShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return fileImportsShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return fileImportsShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return fileImportsShowOptions{}, err
	}

	return fileImportsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildFileImportDetails(resp jsonAPISingleResponse) fileImportDetails {
	resource := resp.Data
	attrs := resource.Attributes
	return fileImportDetails{
		ID:               resource.ID,
		BrokerID:         relationshipIDFromMap(resource.Relationships, "broker"),
		CreatedByID:      relationshipIDFromMap(resource.Relationships, "created-by"),
		FileAttachmentID: relationshipIDFromMap(resource.Relationships, "file-attachment"),
		ProcessedAt:      formatDateTime(stringAttr(attrs, "processed-at")),
		Note:             stringAttr(attrs, "note"),
		RawData:          attrs["raw-data"],
	}
}

func renderFileImportDetails(cmd *cobra.Command, details fileImportDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.BrokerID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	if details.FileAttachmentID != "" {
		fmt.Fprintf(out, "File Attachment: %s\n", details.FileAttachmentID)
	}
	if details.ProcessedAt != "" {
		fmt.Fprintf(out, "Processed At: %s\n", details.ProcessedAt)
	}
	if details.Note != "" {
		fmt.Fprintf(out, "Note: %s\n", details.Note)
	}
	if details.RawData != nil {
		fmt.Fprintln(out, "\nRaw Data:")
		fmt.Fprintln(out, formatJSONBlock(details.RawData, "  "))
	}

	return nil
}
