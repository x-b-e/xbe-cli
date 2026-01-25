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

type commitmentSimulationPeriodsListOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	NoAuth               bool
	Limit                int
	Offset               int
	Sort                 string
	CommitmentSimulation string
}

type commitmentSimulationPeriodRow struct {
	ID                     string `json:"id"`
	Date                   string `json:"date,omitempty"`
	Window                 string `json:"window,omitempty"`
	Iterations             string `json:"iterations,omitempty"`
	Tons                   string `json:"tons,omitempty"`
	CommitmentSimulationID string `json:"commitment_simulation_id,omitempty"`
	CommitmentType         string `json:"commitment_type,omitempty"`
	CommitmentID           string `json:"commitment_id,omitempty"`
}

func newCommitmentSimulationPeriodsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List commitment simulation periods",
		Long: `List commitment simulation periods.

Output Columns:
  ID               Commitment simulation period identifier
  DATE             Period date
  WINDOW           Period window
  ITERATIONS       Iteration count
  TONS             Tons (customer commitments only)
  COMMITMENT SIM   Commitment simulation ID
  COMMITMENT       Commitment type and ID

Filters:
  --commitment-simulation  Filter by commitment simulation ID

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List commitment simulation periods
  xbe view commitment-simulation-periods list

  # Filter by commitment simulation
  xbe view commitment-simulation-periods list --commitment-simulation 123

  # Output as JSON
  xbe view commitment-simulation-periods list --json`,
		Args: cobra.NoArgs,
		RunE: runCommitmentSimulationPeriodsList,
	}
	initCommitmentSimulationPeriodsListFlags(cmd)
	return cmd
}

func init() {
	commitmentSimulationPeriodsCmd.AddCommand(newCommitmentSimulationPeriodsListCmd())
}

func initCommitmentSimulationPeriodsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("commitment-simulation", "", "Filter by commitment simulation ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCommitmentSimulationPeriodsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCommitmentSimulationPeriodsListOptions(cmd)
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
	query.Set("fields[commitment-simulation-periods]", "date,window,iterations,tons,commitment-simulation,commitment")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[commitment_simulation]", opts.CommitmentSimulation)

	body, _, err := client.Get(cmd.Context(), "/v1/commitment-simulation-periods", query)
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

	rows := buildCommitmentSimulationPeriodRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCommitmentSimulationPeriodsTable(cmd, rows)
}

func parseCommitmentSimulationPeriodsListOptions(cmd *cobra.Command) (commitmentSimulationPeriodsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	commitmentSimulation, _ := cmd.Flags().GetString("commitment-simulation")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return commitmentSimulationPeriodsListOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		NoAuth:               noAuth,
		Limit:                limit,
		Offset:               offset,
		Sort:                 sort,
		CommitmentSimulation: commitmentSimulation,
	}, nil
}

func buildCommitmentSimulationPeriodRows(resp jsonAPIResponse) []commitmentSimulationPeriodRow {
	rows := make([]commitmentSimulationPeriodRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := commitmentSimulationPeriodRow{
			ID:         resource.ID,
			Date:       formatDate(stringAttr(resource.Attributes, "date")),
			Window:     stringAttr(resource.Attributes, "window"),
			Iterations: stringAttr(resource.Attributes, "iterations"),
			Tons:       stringAttr(resource.Attributes, "tons"),
		}

		if rel, ok := resource.Relationships["commitment-simulation"]; ok && rel.Data != nil {
			row.CommitmentSimulationID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["commitment"]; ok && rel.Data != nil {
			row.CommitmentType = rel.Data.Type
			row.CommitmentID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderCommitmentSimulationPeriodsTable(cmd *cobra.Command, rows []commitmentSimulationPeriodRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No commitment simulation periods found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tDATE\tWINDOW\tITERATIONS\tTONS\tCOMMITMENT SIM\tCOMMITMENT")
	for _, row := range rows {
		commitment := formatPolymorphic(row.CommitmentType, row.CommitmentID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Date,
			row.Window,
			row.Iterations,
			row.Tons,
			row.CommitmentSimulationID,
			commitment,
		)
	}
	return writer.Flush()
}

func formatPolymorphic(resourceType, resourceID string) string {
	switch {
	case resourceType != "" && resourceID != "":
		return resourceType + "/" + resourceID
	case resourceType != "":
		return resourceType
	default:
		return resourceID
	}
}
