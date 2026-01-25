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

type fileImportsListOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	NoAuth         bool
	Limit          int
	Offset         int
	Sort           string
	Broker         string
	CreatedBy      string
	FileAttachment string
}

type fileImportRow struct {
	ID               string `json:"id"`
	ProcessedAt      string `json:"processed_at,omitempty"`
	Note             string `json:"note,omitempty"`
	BrokerID         string `json:"broker_id,omitempty"`
	CreatedByID      string `json:"created_by_id,omitempty"`
	FileAttachmentID string `json:"file_attachment_id,omitempty"`
}

func newFileImportsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List file imports",
		Long: `List file imports with filtering and pagination.

File imports track uploaded files and their processing status.

Output Columns:
  ID               File import identifier
  PROCESSED AT     When the import was processed
  NOTE             Import note
  BROKER           Broker ID
  CREATED BY       User ID that created the import
  FILE ATTACHMENT  File attachment ID

Filters:
  --broker          Filter by broker ID
  --created-by      Filter by creator user ID
  --file-attachment Filter by file attachment ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List file imports
  xbe view file-imports list

  # Filter by broker
  xbe view file-imports list --broker 123

  # Filter by creator
  xbe view file-imports list --created-by 456

  # Filter by file attachment
  xbe view file-imports list --file-attachment 789

  # Output as JSON
  xbe view file-imports list --json`,
		Args: cobra.NoArgs,
		RunE: runFileImportsList,
	}
	initFileImportsListFlags(cmd)
	return cmd
}

func init() {
	fileImportsCmd.AddCommand(newFileImportsListCmd())
}

func initFileImportsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("created-by", "", "Filter by creator user ID")
	cmd.Flags().String("file-attachment", "", "Filter by file attachment ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runFileImportsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseFileImportsListOptions(cmd)
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
	query.Set("fields[file-imports]", "processed-at,note,broker,created-by,file-attachment")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[file-attachment]", opts.FileAttachment)

	body, _, err := client.Get(cmd.Context(), "/v1/file-imports", query)
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

	rows := buildFileImportRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderFileImportsTable(cmd, rows)
}

func parseFileImportsListOptions(cmd *cobra.Command) (fileImportsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	createdBy, _ := cmd.Flags().GetString("created-by")
	fileAttachment, _ := cmd.Flags().GetString("file-attachment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return fileImportsListOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		NoAuth:         noAuth,
		Limit:          limit,
		Offset:         offset,
		Sort:           sort,
		Broker:         broker,
		CreatedBy:      createdBy,
		FileAttachment: fileAttachment,
	}, nil
}

func buildFileImportRows(resp jsonAPIResponse) []fileImportRow {
	rows := make([]fileImportRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildFileImportRow(resource))
	}
	return rows
}

func buildFileImportRow(resource jsonAPIResource) fileImportRow {
	attrs := resource.Attributes
	return fileImportRow{
		ID:               resource.ID,
		ProcessedAt:      formatDateTime(stringAttr(attrs, "processed-at")),
		Note:             stringAttr(attrs, "note"),
		BrokerID:         relationshipIDFromMap(resource.Relationships, "broker"),
		CreatedByID:      relationshipIDFromMap(resource.Relationships, "created-by"),
		FileAttachmentID: relationshipIDFromMap(resource.Relationships, "file-attachment"),
	}
}

func buildFileImportRowFromSingle(resp jsonAPISingleResponse) fileImportRow {
	return buildFileImportRow(resp.Data)
}

func renderFileImportsTable(cmd *cobra.Command, rows []fileImportRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No file imports found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPROCESSED AT\tNOTE\tBROKER\tCREATED BY\tFILE ATTACHMENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.ProcessedAt,
			truncateString(row.Note, 30),
			row.BrokerID,
			row.CreatedByID,
			row.FileAttachmentID,
		)
	}
	return writer.Flush()
}
