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

type commitmentSimulationsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	Commitment   string
	Status       string
	CreatedAtMin string
	CreatedAtMax string
	UpdatedAtMin string
	UpdatedAtMax string
}

type commitmentSimulationRow struct {
	ID             string `json:"id"`
	CommitmentType string `json:"commitment_type,omitempty"`
	CommitmentID   string `json:"commitment_id,omitempty"`
	StartOn        string `json:"start_on,omitempty"`
	EndOn          string `json:"end_on,omitempty"`
	IterationCount int    `json:"iteration_count,omitempty"`
	Status         string `json:"status,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
}

func newCommitmentSimulationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List commitment simulations",
		Long: `List commitment simulations.

Output Columns:
  ID           Simulation identifier
  COMMITMENT   Commitment type and ID
  START ON     Simulation start date
  END ON       Simulation end date
  ITERATIONS   Iteration count
  STATUS       Simulation status
  CREATED AT   Creation timestamp

Filters:
  --commitment      Filter by commitment ID
  --status          Filter by status (enqueued, processed)
  --created-at-min  Filter by created-at on/after (ISO 8601)
  --created-at-max  Filter by created-at on/before (ISO 8601)
  --updated-at-min  Filter by updated-at on/after (ISO 8601)
  --updated-at-max  Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List commitment simulations
  xbe view commitment-simulations list

  # Filter by commitment and status
  xbe view commitment-simulations list --commitment 123 --status enqueued

  # Output as JSON
  xbe view commitment-simulations list --json`,
		Args: cobra.NoArgs,
		RunE: runCommitmentSimulationsList,
	}
	initCommitmentSimulationsListFlags(cmd)
	return cmd
}

func init() {
	commitmentSimulationsCmd.AddCommand(newCommitmentSimulationsListCmd())
}

func initCommitmentSimulationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("commitment", "", "Filter by commitment ID")
	cmd.Flags().String("status", "", "Filter by status (enqueued, processed)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCommitmentSimulationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCommitmentSimulationsListOptions(cmd)
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
	query.Set("fields[commitment-simulations]", "commitment,start-on,end-on,iteration-count,status,created-at")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[commitment]", opts.Commitment)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/commitment-simulations", query)
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

	rows := buildCommitmentSimulationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCommitmentSimulationsTable(cmd, rows)
}

func parseCommitmentSimulationsListOptions(cmd *cobra.Command) (commitmentSimulationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	commitment, _ := cmd.Flags().GetString("commitment")
	status, _ := cmd.Flags().GetString("status")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return commitmentSimulationsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		Commitment:   commitment,
		Status:       status,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
	}, nil
}

func buildCommitmentSimulationRows(resp jsonAPIResponse) []commitmentSimulationRow {
	rows := make([]commitmentSimulationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildCommitmentSimulationRow(resource))
	}
	return rows
}

func buildCommitmentSimulationRow(resource jsonAPIResource) commitmentSimulationRow {
	row := commitmentSimulationRow{
		ID:             resource.ID,
		StartOn:        formatDate(stringAttr(resource.Attributes, "start-on")),
		EndOn:          formatDate(stringAttr(resource.Attributes, "end-on")),
		IterationCount: intAttr(resource.Attributes, "iteration-count"),
		Status:         stringAttr(resource.Attributes, "status"),
		CreatedAt:      formatDateTime(stringAttr(resource.Attributes, "created-at")),
	}

	if rel, ok := resource.Relationships["commitment"]; ok && rel.Data != nil {
		row.CommitmentType = rel.Data.Type
		row.CommitmentID = rel.Data.ID
	}

	return row
}

func buildCommitmentSimulationRowFromSingle(resp jsonAPISingleResponse) commitmentSimulationRow {
	return buildCommitmentSimulationRow(resp.Data)
}

func renderCommitmentSimulationsTable(cmd *cobra.Command, rows []commitmentSimulationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No commitment simulations found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tCOMMITMENT\tSTART ON\tEND ON\tITERATIONS\tSTATUS\tCREATED AT")
	for _, row := range rows {
		commitment := ""
		if row.CommitmentType != "" && row.CommitmentID != "" {
			commitment = row.CommitmentType + "/" + row.CommitmentID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
			row.ID,
			truncateString(commitment, 32),
			row.StartOn,
			row.EndOn,
			row.IterationCount,
			row.Status,
			row.CreatedAt,
		)
	}
	return writer.Flush()
}
