package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type incidentRequestApprovalsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

type incidentRequestApprovalRow struct {
	ID                string `json:"id"`
	IncidentRequestID string `json:"incident_request_id,omitempty"`
	Comment           string `json:"comment,omitempty"`
}

func newIncidentRequestApprovalsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List incident request approvals",
		Long: `List incident request approvals.

Output Columns:
  ID                Approval identifier
  INCIDENT REQUEST  Incident request ID
  COMMENT           Comment (truncated)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List incident request approvals
  xbe view incident-request-approvals list

  # JSON output
  xbe view incident-request-approvals list --json`,
		Args: cobra.NoArgs,
		RunE: runIncidentRequestApprovalsList,
	}
	initIncidentRequestApprovalsListFlags(cmd)
	return cmd
}

func init() {
	incidentRequestApprovalsCmd.AddCommand(newIncidentRequestApprovalsListCmd())
}

func initIncidentRequestApprovalsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runIncidentRequestApprovalsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseIncidentRequestApprovalsListOptions(cmd)
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
	query.Set("fields[incident-request-approvals]", "incident-request,comment")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	body, status, err := client.Get(cmd.Context(), "/v1/incident-request-approvals", query)
	if err != nil {
		if status == http.StatusNotFound {
			return renderIncidentRequestApprovalsUnavailable(cmd, opts.JSON)
		}
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

	rows := buildIncidentRequestApprovalRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderIncidentRequestApprovalsTable(cmd, rows)
}

func renderIncidentRequestApprovalsUnavailable(cmd *cobra.Command, jsonOut bool) error {
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), []incidentRequestApprovalRow{})
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Incident request approvals are write-only; list is not available.")
	return nil
}

func parseIncidentRequestApprovalsListOptions(cmd *cobra.Command) (incidentRequestApprovalsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return incidentRequestApprovalsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildIncidentRequestApprovalRows(resp jsonAPIResponse) []incidentRequestApprovalRow {
	rows := make([]incidentRequestApprovalRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildIncidentRequestApprovalRow(resource)
		rows = append(rows, row)
	}
	return rows
}

func buildIncidentRequestApprovalRow(resource jsonAPIResource) incidentRequestApprovalRow {
	attrs := resource.Attributes
	row := incidentRequestApprovalRow{
		ID:      resource.ID,
		Comment: strings.TrimSpace(stringAttr(attrs, "comment")),
	}

	if rel, ok := resource.Relationships["incident-request"]; ok && rel.Data != nil {
		row.IncidentRequestID = rel.Data.ID
	}

	return row
}

func renderIncidentRequestApprovalsTable(cmd *cobra.Command, rows []incidentRequestApprovalRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No incident request approvals found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tINCIDENT REQUEST\tCOMMENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			row.IncidentRequestID,
			truncateString(row.Comment, 40),
		)
	}
	return writer.Flush()
}
