package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type projectImportFileVerificationsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Project string
}

type projectImportFileVerificationRow struct {
	ID                  string `json:"id"`
	VerificationType    string `json:"verification_type,omitempty"`
	Status              string `json:"status,omitempty"`
	VerificationSummary string `json:"verification_summary,omitempty"`
	ItemsMismatching    string `json:"items_mismatching,omitempty"`
	IsDryRun            bool   `json:"is_dry_run"`
	ProjectID           string `json:"project_id,omitempty"`
	FileImportID        string `json:"file_import_id,omitempty"`
	ProcessedAt         string `json:"processed_at,omitempty"`
	ExtractedJobNumber  string `json:"extracted_job_number,omitempty"`
	ExternalRecordID    string `json:"external_record_id,omitempty"`
}

func newProjectImportFileVerificationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project import file verifications",
		Long: `List project import file verifications.

Project import file verifications require a project filter. Results include
status and verification summary details.

Output Columns:
  ID            Verification identifier
  TYPE          Verification type
  STATUS        Verification status
  SUMMARY       Verification summary (if present)
  MISMATCHING   Items mismatching (if present)
  DRY RUN       Whether the verification was a dry run
  PROJECT       Project ID
  FILE IMPORT   File import ID
  PROCESSED AT  Processing timestamp (if present)

Filters:
  --project  Project ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # List verifications for a project
  xbe view project-import-file-verifications list --project 123

  # Output as JSON
  xbe view project-import-file-verifications list --project 123 --json`,
		Args: cobra.NoArgs,
		RunE: runProjectImportFileVerificationsList,
	}
	initProjectImportFileVerificationsListFlags(cmd)
	return cmd
}

func init() {
	projectImportFileVerificationsCmd.AddCommand(newProjectImportFileVerificationsListCmd())
}

func initProjectImportFileVerificationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("project", "", "Project ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectImportFileVerificationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectImportFileVerificationsListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Project) == "" {
		err := fmt.Errorf("--project is required")
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
	query.Set("fields[project-import-file-verifications]", "verification-type,is-dry-run,status,verification-summary,items-mismatching,processed-at,extracted-job-number,external-record-id,project,file-import")
	query.Set("filter[project]", opts.Project)

	body, _, err := client.Get(cmd.Context(), "/v1/project-import-file-verifications", query)
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

	rows := buildProjectImportFileVerificationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectImportFileVerificationsTable(cmd, rows)
}

func parseProjectImportFileVerificationsListOptions(cmd *cobra.Command) (projectImportFileVerificationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	project, _ := cmd.Flags().GetString("project")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectImportFileVerificationsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Project: project,
	}, nil
}

func buildProjectImportFileVerificationRows(resp jsonAPIResponse) []projectImportFileVerificationRow {
	rows := make([]projectImportFileVerificationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildProjectImportFileVerificationRow(resource))
	}
	return rows
}

func buildProjectImportFileVerificationRow(resource jsonAPIResource) projectImportFileVerificationRow {
	attrs := resource.Attributes
	row := projectImportFileVerificationRow{
		ID:                  resource.ID,
		VerificationType:    stringAttr(attrs, "verification-type"),
		Status:              stringAttr(attrs, "status"),
		VerificationSummary: stringAttr(attrs, "verification-summary"),
		ItemsMismatching:    stringAttr(attrs, "items-mismatching"),
		IsDryRun:            boolAttr(attrs, "is-dry-run"),
		ProcessedAt:         formatDateTime(stringAttr(attrs, "processed-at")),
		ExtractedJobNumber:  stringAttr(attrs, "extracted-job-number"),
		ExternalRecordID:    stringAttr(attrs, "external-record-id"),
	}

	if rel, ok := resource.Relationships["project"]; ok && rel.Data != nil {
		row.ProjectID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["file-import"]; ok && rel.Data != nil {
		row.FileImportID = rel.Data.ID
	}

	return row
}

func renderProjectImportFileVerificationsTable(cmd *cobra.Command, rows []projectImportFileVerificationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project import file verifications found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTYPE\tSTATUS\tSUMMARY\tMISMATCHING\tDRY RUN\tPROJECT\tFILE IMPORT\tPROCESSED AT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%t\t%s\t%s\t%s\n",
			row.ID,
			row.VerificationType,
			row.Status,
			truncateString(row.VerificationSummary, 50),
			truncateString(row.ItemsMismatching, 40),
			row.IsDryRun,
			row.ProjectID,
			row.FileImportID,
			row.ProcessedAt,
		)
	}
	return writer.Flush()
}
