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

type commitmentItemsListOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	NoAuth         bool
	Limit          int
	Offset         int
	Sort           string
	CommitmentType string
	CommitmentID   string
	Status         string
}

type commitmentItemRow struct {
	ID                         string `json:"id"`
	Label                      string `json:"label,omitempty"`
	Status                     string `json:"status,omitempty"`
	StartOn                    string `json:"start_on,omitempty"`
	EndOn                      string `json:"end_on,omitempty"`
	AdjustmentSequencePosition string `json:"adjustment_sequence_position,omitempty"`
	CommitmentType             string `json:"commitment_type,omitempty"`
	CommitmentID               string `json:"commitment_id,omitempty"`
}

func newCommitmentItemsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List commitment items",
		Long: `List commitment items.

Output Columns:
  ID           Commitment item identifier
  LABEL        Commitment item label
  STATUS       Commitment item status
  START        Start date
  END          End date
  SEQ          Adjustment sequence position
  COMMITMENT   Commitment type and ID

Filters:
  --commitment-type  Filter by commitment type (requires --commitment-id)
  --commitment-id    Filter by commitment ID (requires --commitment-type)
  --status           Filter by status (editing, active, inactive)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List commitment items
  xbe view commitment-items list

  # Filter by status
  xbe view commitment-items list --status active

  # Filter by commitment
  xbe view commitment-items list --commitment-type customer-commitments --commitment-id 123

  # Output as JSON
  xbe view commitment-items list --json`,
		Args: cobra.NoArgs,
		RunE: runCommitmentItemsList,
	}
	initCommitmentItemsListFlags(cmd)
	return cmd
}

func init() {
	commitmentItemsCmd.AddCommand(newCommitmentItemsListCmd())
}

func initCommitmentItemsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("commitment-type", "", "Filter by commitment type")
	cmd.Flags().String("commitment-id", "", "Filter by commitment ID")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCommitmentItemsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCommitmentItemsListOptions(cmd)
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
	query.Set("fields[commitment-items]", "label,status,start-on,end-on,adjustment-sequence-position,commitment")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[status]", opts.Status)
	if opts.CommitmentType != "" && opts.CommitmentID != "" {
		query.Set("filter[commitment]", opts.CommitmentType+"|"+opts.CommitmentID)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/commitment-items", query)
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

	rows := buildCommitmentItemRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCommitmentItemsTable(cmd, rows)
}

func parseCommitmentItemsListOptions(cmd *cobra.Command) (commitmentItemsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	commitmentType, _ := cmd.Flags().GetString("commitment-type")
	commitmentID, _ := cmd.Flags().GetString("commitment-id")
	status, _ := cmd.Flags().GetString("status")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return commitmentItemsListOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		NoAuth:         noAuth,
		Limit:          limit,
		Offset:         offset,
		Sort:           sort,
		CommitmentType: commitmentType,
		CommitmentID:   commitmentID,
		Status:         status,
	}, nil
}

func buildCommitmentItemRows(resp jsonAPIResponse) []commitmentItemRow {
	rows := make([]commitmentItemRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildCommitmentItemRow(resource))
	}
	return rows
}

func buildCommitmentItemRow(resource jsonAPIResource) commitmentItemRow {
	row := commitmentItemRow{
		ID:                         resource.ID,
		Label:                      strings.TrimSpace(stringAttr(resource.Attributes, "label")),
		Status:                     stringAttr(resource.Attributes, "status"),
		StartOn:                    formatDate(stringAttr(resource.Attributes, "start-on")),
		EndOn:                      formatDate(stringAttr(resource.Attributes, "end-on")),
		AdjustmentSequencePosition: stringAttr(resource.Attributes, "adjustment-sequence-position"),
	}

	if rel, ok := resource.Relationships["commitment"]; ok && rel.Data != nil {
		row.CommitmentType = rel.Data.Type
		row.CommitmentID = rel.Data.ID
	}

	return row
}

func buildCommitmentItemRowFromSingle(resp jsonAPISingleResponse) commitmentItemRow {
	return buildCommitmentItemRow(resp.Data)
}

func renderCommitmentItemsTable(cmd *cobra.Command, rows []commitmentItemRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No commitment items found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tLABEL\tSTATUS\tSTART\tEND\tSEQ\tCOMMITMENT")
	for _, row := range rows {
		commitment := ""
		if row.CommitmentType != "" {
			commitment = row.CommitmentType
		}
		if row.CommitmentID != "" {
			if commitment != "" {
				commitment += "/"
			}
			commitment += row.CommitmentID
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Label, 26),
			row.Status,
			row.StartOn,
			row.EndOn,
			row.AdjustmentSequencePosition,
			truncateString(commitment, 32),
		)
	}
	return writer.Flush()
}
