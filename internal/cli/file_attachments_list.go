package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type fileAttachmentsListOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	NoAuth         bool
	Limit          int
	Offset         int
	AttachedToType string
	AttachedToID   string
	CreatedBy      string
}

type fileAttachmentRow struct {
	ID             string `json:"id"`
	FileName       string `json:"file_name"`
	ObjectKey      string `json:"object_key"`
	CanOptimize    bool   `json:"can_optimize"`
	AttachedToType string `json:"attached_to_type,omitempty"`
	AttachedToID   string `json:"attached_to_id,omitempty"`
	CreatedByID    string `json:"created_by_id,omitempty"`
}

func newFileAttachmentsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List file attachments",
		Long: `List file attachments with filtering and pagination.

Output Columns:
  ID           File attachment identifier
  FILE NAME    File name
  OBJECT KEY   S3 object key
  ATTACHED TO  Attached resource (type/id)
  CREATED BY   User who created the attachment

Filters:
  --attached-to-type  Filter by attached resource type (use with --attached-to-id, e.g., Project)
  --attached-to-id    Filter by attached resource ID (use with --attached-to-type)
  --created-by        Filter by creator user ID`,
		Example: `  # List file attachments
  xbe view file-attachments list

  # Filter by attached resource
  xbe view file-attachments list --attached-to-type Project --attached-to-id 123

  # Filter by creator
  xbe view file-attachments list --created-by 456

  # Paginate results
  xbe view file-attachments list --limit 20 --offset 40

  # Output as JSON
  xbe view file-attachments list --json`,
		RunE: runFileAttachmentsList,
	}
	initFileAttachmentsListFlags(cmd)
	return cmd
}

func init() {
	fileAttachmentsCmd.AddCommand(newFileAttachmentsListCmd())
}

func initFileAttachmentsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("attached-to-type", "", "Filter by attached resource type (e.g., Project)")
	cmd.Flags().String("attached-to-id", "", "Filter by attached resource ID")
	cmd.Flags().String("created-by", "", "Filter by creator user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runFileAttachmentsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseFileAttachmentsListOptions(cmd)
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[file-attachments]", "file-name,object-key,can-optimize,attached-to,created-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	if opts.AttachedToType != "" && opts.AttachedToID != "" {
		query.Set("filter[by_attached_to]", opts.AttachedToID+"|"+opts.AttachedToType)
	}
	setFilterIfPresent(query, "filter[created_by]", opts.CreatedBy)

	body, _, err := client.Get(cmd.Context(), "/v1/file-attachments", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	rows := buildFileAttachmentRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderFileAttachmentsTable(cmd, rows)
}

func parseFileAttachmentsListOptions(cmd *cobra.Command) (fileAttachmentsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	attachedToType, _ := cmd.Flags().GetString("attached-to-type")
	attachedToID, _ := cmd.Flags().GetString("attached-to-id")
	createdBy, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return fileAttachmentsListOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		NoAuth:         noAuth,
		Limit:          limit,
		Offset:         offset,
		AttachedToType: attachedToType,
		AttachedToID:   attachedToID,
		CreatedBy:      createdBy,
	}, nil
}

func buildFileAttachmentRows(resp jsonAPIResponse) []fileAttachmentRow {
	rows := make([]fileAttachmentRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := fileAttachmentRow{
			ID:          resource.ID,
			FileName:    stringAttr(resource.Attributes, "file-name"),
			ObjectKey:   stringAttr(resource.Attributes, "object-key"),
			CanOptimize: boolAttr(resource.Attributes, "can-optimize"),
		}

		if rel, ok := resource.Relationships["attached-to"]; ok && rel.Data != nil {
			row.AttachedToType = rel.Data.Type
			row.AttachedToID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
			row.CreatedByID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderFileAttachmentsTable(cmd *cobra.Command, rows []fileAttachmentRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No file attachments found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tFILE NAME\tOBJECT KEY\tATTACHED TO\tCREATED BY")
	for _, row := range rows {
		attachedTo := ""
		if row.AttachedToType != "" && row.AttachedToID != "" {
			attachedTo = row.AttachedToType + "/" + row.AttachedToID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.FileName, 30),
			truncateString(row.ObjectKey, 30),
			truncateString(attachedTo, 30),
			row.CreatedByID,
		)
	}
	return writer.Flush()
}
