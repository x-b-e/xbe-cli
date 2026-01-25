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

type projectsFileImportsListOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	NoAuth         bool
	Limit          int
	Offset         int
	Sort           string
	FileImportType string
	IsDryRun       string
	FileImport     string
	SubjectType    string
	SubjectID      string
}

type projectsFileImportRow struct {
	ID             string `json:"id"`
	FileImportType string `json:"file_import_type,omitempty"`
	IsDryRun       bool   `json:"is_dry_run,omitempty"`
	Status         string `json:"status,omitempty"`
	FileImportID   string `json:"file_import_id,omitempty"`
	SubjectType    string `json:"subject_type,omitempty"`
	SubjectID      string `json:"subject_id,omitempty"`
}

func newProjectsFileImportsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects file imports",
		Long: `List projects file imports.

Output Columns:
  ID          Import identifier
  TYPE        File import type
  FILE IMPORT File import ID
  SUBJECT     Subject type and ID
  STATUS      Processing status
  DRY RUN     Whether the import is a dry run

Filters:
  --file-import-type  Filter by file import type
  --is-dry-run        Filter by dry run status (true/false)
  --file-import       Filter by file import ID
  --subject-type      Filter by subject type (use with --subject-id)
  --subject-id        Filter by subject ID (use with --subject-type)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List projects file imports
  xbe view projects-file-imports list

  # Filter by file import type
  xbe view projects-file-imports list --file-import-type SageProjectsFileImport

  # Filter by subject
  xbe view projects-file-imports list --subject-type brokers --subject-id 456

  # JSON output
  xbe view projects-file-imports list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectsFileImportsList,
	}
	initProjectsFileImportsListFlags(cmd)
	return cmd
}

func init() {
	projectsFileImportsCmd.AddCommand(newProjectsFileImportsListCmd())
}

func initProjectsFileImportsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Limit results")
	cmd.Flags().Int("offset", 0, "Offset results")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("file-import-type", "", "Filter by file import type")
	cmd.Flags().String("is-dry-run", "", "Filter by dry run status (true/false)")
	cmd.Flags().String("file-import", "", "Filter by file import ID")
	cmd.Flags().String("subject-type", "", "Filter by subject type")
	cmd.Flags().String("subject-id", "", "Filter by subject ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectsFileImportsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectsFileImportsListOptions(cmd)
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
	query.Set("fields[projects-file-imports]", "file-import-type,is-dry-run,status,file-import,subject")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[file-import-type]", opts.FileImportType)
	setFilterIfPresent(query, "filter[is-dry-run]", opts.IsDryRun)
	setFilterIfPresent(query, "filter[file-import]", opts.FileImport)

	if opts.SubjectType != "" && opts.SubjectID != "" {
		query.Set("filter[subject]", opts.SubjectType+"|"+opts.SubjectID)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/projects-file-imports", query)
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

	rows := buildProjectsFileImportRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectsFileImportsTable(cmd, rows)
}

func parseProjectsFileImportsListOptions(cmd *cobra.Command) (projectsFileImportsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	fileImportType, _ := cmd.Flags().GetString("file-import-type")
	isDryRun, _ := cmd.Flags().GetString("is-dry-run")
	fileImport, _ := cmd.Flags().GetString("file-import")
	subjectType, _ := cmd.Flags().GetString("subject-type")
	subjectID, _ := cmd.Flags().GetString("subject-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectsFileImportsListOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		NoAuth:         noAuth,
		Limit:          limit,
		Offset:         offset,
		Sort:           sort,
		FileImportType: fileImportType,
		IsDryRun:       isDryRun,
		FileImport:     fileImport,
		SubjectType:    subjectType,
		SubjectID:      subjectID,
	}, nil
}

func buildProjectsFileImportRows(resp jsonAPIResponse) []projectsFileImportRow {
	rows := make([]projectsFileImportRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := projectsFileImportRow{
			ID:             resource.ID,
			FileImportType: stringAttr(attrs, "file-import-type"),
			IsDryRun:       boolAttr(attrs, "is-dry-run"),
			Status:         stringAttr(attrs, "status"),
		}

		if rel, ok := resource.Relationships["file-import"]; ok && rel.Data != nil {
			row.FileImportID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["subject"]; ok && rel.Data != nil {
			row.SubjectType = rel.Data.Type
			row.SubjectID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderProjectsFileImportsTable(cmd *cobra.Command, rows []projectsFileImportRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No projects file imports found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTYPE\tFILE IMPORT\tSUBJECT\tSTATUS\tDRY RUN")
	for _, row := range rows {
		subject := ""
		if row.SubjectType != "" && row.SubjectID != "" {
			subject = row.SubjectType + "/" + row.SubjectID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.FileImportType, 30),
			row.FileImportID,
			truncateString(subject, 30),
			truncateString(row.Status, 20),
			formatBool(row.IsDryRun),
		)
	}
	return writer.Flush()
}
