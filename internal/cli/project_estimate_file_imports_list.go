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

type projectEstimateFileImportsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

type projectEstimateFileImportRow struct {
	ID                         string `json:"id"`
	FileImportType             string `json:"file_import_type,omitempty"`
	IsDryRun                   bool   `json:"is_dry_run,omitempty"`
	ShouldUpdateFileExtraction bool   `json:"should_update_file_extraction,omitempty"`
	FileImportID               string `json:"file_import_id,omitempty"`
	ProjectID                  string `json:"project_id,omitempty"`
}

func newProjectEstimateFileImportsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project estimate file imports",
		Long: `List project estimate file imports.

Project estimate file imports are executed on demand and are not persisted, so
this list is typically empty.

Output Columns:
  ID                Import identifier
  TYPE              File import type
  FILE IMPORT        File import ID
  PROJECT            Project ID
  DRY RUN            Whether the import is a dry run
  UPDATE EXTRACTION  Whether PDF extraction data should be updated

Filters:
  None

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List project estimate file imports
  xbe view project-estimate-file-imports list

  # Paginate results
  xbe view project-estimate-file-imports list --limit 10 --offset 20

  # JSON output
  xbe view project-estimate-file-imports list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectEstimateFileImportsList,
	}
	initProjectEstimateFileImportsListFlags(cmd)
	return cmd
}

func init() {
	projectEstimateFileImportsCmd.AddCommand(newProjectEstimateFileImportsListCmd())
}

func initProjectEstimateFileImportsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Limit results")
	cmd.Flags().Int("offset", 0, "Offset results")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectEstimateFileImportsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectEstimateFileImportsListOptions(cmd)
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
	query.Set("fields[project-estimate-file-imports]", "file-import-type,is-dry-run,should-update-file-extraction,file-import,project")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/project-estimate-file-imports", query)
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

	rows := buildProjectEstimateFileImportRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectEstimateFileImportsTable(cmd, rows)
}

func parseProjectEstimateFileImportsListOptions(cmd *cobra.Command) (projectEstimateFileImportsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectEstimateFileImportsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildProjectEstimateFileImportRows(resp jsonAPIResponse) []projectEstimateFileImportRow {
	rows := make([]projectEstimateFileImportRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := projectEstimateFileImportRow{
			ID:                         resource.ID,
			FileImportType:             stringAttr(attrs, "file-import-type"),
			IsDryRun:                   boolAttr(attrs, "is-dry-run"),
			ShouldUpdateFileExtraction: boolAttr(attrs, "should-update-file-extraction"),
		}

		if rel, ok := resource.Relationships["file-import"]; ok && rel.Data != nil {
			row.FileImportID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["project"]; ok && rel.Data != nil {
			row.ProjectID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderProjectEstimateFileImportsTable(cmd *cobra.Command, rows []projectEstimateFileImportRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project estimate file imports found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTYPE\tFILE IMPORT\tPROJECT\tDRY RUN\tUPDATE EXTRACTION")
	for _, row := range rows {
		dryRun := "no"
		if row.IsDryRun {
			dryRun = "yes"
		}
		update := "no"
		if row.ShouldUpdateFileExtraction {
			update = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.FileImportType, 30),
			row.FileImportID,
			row.ProjectID,
			dryRun,
			update,
		)
	}
	return writer.Flush()
}
