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

type incidentRequestRejectionsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

type incidentRequestRejectionRow struct {
	ID                string `json:"id"`
	IncidentRequestID string `json:"incident_request_id,omitempty"`
	Comment           string `json:"comment,omitempty"`
}

func newIncidentRequestRejectionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List incident request rejections",
		Long: `List incident request rejections.

Output Columns:
  ID                Rejection identifier
  INCIDENT REQUEST  Incident request ID
  COMMENT           Comment (truncated)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List incident request rejections
  xbe view incident-request-rejections list

  # JSON output
  xbe view incident-request-rejections list --json`,
		Args: cobra.NoArgs,
		RunE: runIncidentRequestRejectionsList,
	}
	initIncidentRequestRejectionsListFlags(cmd)
	return cmd
}

func init() {
	incidentRequestRejectionsCmd.AddCommand(newIncidentRequestRejectionsListCmd())
}

func initIncidentRequestRejectionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runIncidentRequestRejectionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseIncidentRequestRejectionsListOptions(cmd)
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
	query.Set("fields[incident-request-rejections]", "incident-request,comment")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	body, status, err := client.Get(cmd.Context(), "/v1/incident-request-rejections", query)
	if err != nil {
		if status == http.StatusNotFound {
			return renderIncidentRequestRejectionsUnavailable(cmd, opts.JSON)
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

	rows := buildIncidentRequestRejectionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderIncidentRequestRejectionsTable(cmd, rows)
}

func renderIncidentRequestRejectionsUnavailable(cmd *cobra.Command, jsonOut bool) error {
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), []incidentRequestRejectionRow{})
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Incident request rejections are write-only; list is not available.")
	return nil
}

func parseIncidentRequestRejectionsListOptions(cmd *cobra.Command) (incidentRequestRejectionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return incidentRequestRejectionsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildIncidentRequestRejectionRows(resp jsonAPIResponse) []incidentRequestRejectionRow {
	rows := make([]incidentRequestRejectionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildIncidentRequestRejectionRow(resource)
		rows = append(rows, row)
	}
	return rows
}

func buildIncidentRequestRejectionRow(resource jsonAPIResource) incidentRequestRejectionRow {
	attrs := resource.Attributes
	row := incidentRequestRejectionRow{
		ID:      resource.ID,
		Comment: strings.TrimSpace(stringAttr(attrs, "comment")),
	}

	if rel, ok := resource.Relationships["incident-request"]; ok && rel.Data != nil {
		row.IncidentRequestID = rel.Data.ID
	}

	return row
}

func renderIncidentRequestRejectionsTable(cmd *cobra.Command, rows []incidentRequestRejectionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No incident request rejections found.")
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
