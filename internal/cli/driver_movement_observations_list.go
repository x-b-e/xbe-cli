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

type driverMovementObservationsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	Plan         string
	IsCurrent    bool
	CreatedAtMin string
	CreatedAtMax string
	UpdatedAtMin string
	UpdatedAtMax string
}

type driverMovementObservationRow struct {
	ID        string `json:"id"`
	PlanID    string `json:"plan_id,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

func newDriverMovementObservationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List driver movement observations",
		Long: `List driver movement observations.

Output Columns:
  ID         Observation identifier
  PLAN       Job production plan ID
  CREATED    Created timestamp
  UPDATED    Updated timestamp

Filters:
  --plan             Filter by job production plan ID
  --is-current       Only include current observations
  --created-at-min   Filter by created-at on/after (ISO 8601)
  --created-at-max   Filter by created-at on/before (ISO 8601)
  --updated-at-min   Filter by updated-at on/after (ISO 8601)
  --updated-at-max   Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List observations
  xbe view driver-movement-observations list

  # Filter by plan
  xbe view driver-movement-observations list --plan 123

  # Only current observations
  xbe view driver-movement-observations list --is-current

  # Output as JSON
  xbe view driver-movement-observations list --json`,
		Args: cobra.NoArgs,
		RunE: runDriverMovementObservationsList,
	}
	initDriverMovementObservationsListFlags(cmd)
	return cmd
}

func init() {
	driverMovementObservationsCmd.AddCommand(newDriverMovementObservationsListCmd())
}

func initDriverMovementObservationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("plan", "", "Filter by job production plan ID")
	cmd.Flags().Bool("is-current", false, "Only include current observations")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDriverMovementObservationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseDriverMovementObservationsListOptions(cmd)
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
	query.Set("fields[driver-movement-observations]", "created-at,updated-at,plan")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[plan]", opts.Plan)
	if opts.IsCurrent {
		query.Set("filter[is-current]", "true")
	}
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/driver-movement-observations", query)
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

	rows := buildDriverMovementObservationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderDriverMovementObservationsTable(cmd, rows)
}

func parseDriverMovementObservationsListOptions(cmd *cobra.Command) (driverMovementObservationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	plan, _ := cmd.Flags().GetString("plan")
	isCurrent, _ := cmd.Flags().GetBool("is-current")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return driverMovementObservationsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		Plan:         plan,
		IsCurrent:    isCurrent,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
	}, nil
}

func buildDriverMovementObservationRows(resp jsonAPIResponse) []driverMovementObservationRow {
	rows := make([]driverMovementObservationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildDriverMovementObservationRow(resource))
	}
	return rows
}

func buildDriverMovementObservationRow(resource jsonAPIResource) driverMovementObservationRow {
	attrs := resource.Attributes
	row := driverMovementObservationRow{
		ID:        resource.ID,
		CreatedAt: strings.TrimSpace(stringAttr(attrs, "created-at")),
		UpdatedAt: strings.TrimSpace(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["plan"]; ok && rel.Data != nil {
		row.PlanID = rel.Data.ID
	}

	return row
}

func renderDriverMovementObservationsTable(cmd *cobra.Command, rows []driverMovementObservationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No driver movement observations found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPLAN\tCREATED\tUPDATED")
	for _, row := range rows {
		createdAt := formatDate(row.CreatedAt)
		updatedAt := formatDate(row.UpdatedAt)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			row.PlanID,
			createdAt,
			updatedAt,
		)
	}
	return writer.Flush()
}
