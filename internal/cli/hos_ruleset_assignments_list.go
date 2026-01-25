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

type hosRulesetAssignmentsListOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	NoAuth         bool
	Limit          int
	Offset         int
	Sort           string
	Broker         string
	HosDay         string
	User           string
	Driver         string
	EffectiveAtMin string
	EffectiveAtMax string
	IsEffectiveAt  string
}

type hosRulesetAssignmentRow struct {
	ID          string `json:"id"`
	RuleSetID   string `json:"rule_set_id,omitempty"`
	Name        string `json:"name,omitempty"`
	EffectiveAt string `json:"effective_at,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	HosDayID    string `json:"hos_day_id,omitempty"`
	BrokerID    string `json:"broker_id,omitempty"`
}

func newHosRulesetAssignmentsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List HOS ruleset assignments",
		Long: `List HOS ruleset assignments.

Output Columns:
  ID            HOS ruleset assignment identifier
  RULE SET      Rule set identifier
  NAME          Rule set name
  EFFECTIVE AT  When the rule set became active
  USER          Driver (user) ID
  HOS DAY       HOS day ID
  BROKER        Broker ID

Filters:
  --broker            Filter by broker ID
  --hos-day           Filter by HOS day ID
  --user              Filter by user (driver) ID
  --driver            Filter by driver ID (alias for --user)
  --effective-at-min  Filter by effective-at on/after (ISO 8601)
  --effective-at-max  Filter by effective-at on/before (ISO 8601)
  --is-effective-at   Filter by presence of effective-at (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List HOS ruleset assignments
  xbe view hos-ruleset-assignments list

  # Filter by driver
  xbe view hos-ruleset-assignments list --driver 123

  # Filter by effective-at range
  xbe view hos-ruleset-assignments list --effective-at-min 2025-01-01T00:00:00Z --effective-at-max 2025-01-31T23:59:59Z

  # Output as JSON
  xbe view hos-ruleset-assignments list --json`,
		Args: cobra.NoArgs,
		RunE: runHosRulesetAssignmentsList,
	}
	initHosRulesetAssignmentsListFlags(cmd)
	return cmd
}

func init() {
	hosRulesetAssignmentsCmd.AddCommand(newHosRulesetAssignmentsListCmd())
}

func initHosRulesetAssignmentsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("hos-day", "", "Filter by HOS day ID")
	cmd.Flags().String("user", "", "Filter by user (driver) ID")
	cmd.Flags().String("driver", "", "Filter by driver ID (alias for --user)")
	cmd.Flags().String("effective-at-min", "", "Filter by effective-at on/after (ISO 8601)")
	cmd.Flags().String("effective-at-max", "", "Filter by effective-at on/before (ISO 8601)")
	cmd.Flags().String("is-effective-at", "", "Filter by presence of effective-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runHosRulesetAssignmentsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseHosRulesetAssignmentsListOptions(cmd)
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
	query.Set("fields[hos-ruleset-assignments]", "rule-set-id,name,effective-at,broker,hos-day,user")

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
	setFilterIfPresent(query, "filter[hos-day]", opts.HosDay)
	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[driver]", opts.Driver)
	setFilterIfPresent(query, "filter[effective-at-min]", opts.EffectiveAtMin)
	setFilterIfPresent(query, "filter[effective-at-max]", opts.EffectiveAtMax)
	setFilterIfPresent(query, "filter[is-effective-at]", opts.IsEffectiveAt)

	body, _, err := client.Get(cmd.Context(), "/v1/hos-ruleset-assignments", query)
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

	rows := buildHosRulesetAssignmentRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderHosRulesetAssignmentsTable(cmd, rows)
}

func parseHosRulesetAssignmentsListOptions(cmd *cobra.Command) (hosRulesetAssignmentsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	hosDay, _ := cmd.Flags().GetString("hos-day")
	user, _ := cmd.Flags().GetString("user")
	driver, _ := cmd.Flags().GetString("driver")
	effectiveAtMin, _ := cmd.Flags().GetString("effective-at-min")
	effectiveAtMax, _ := cmd.Flags().GetString("effective-at-max")
	isEffectiveAt, _ := cmd.Flags().GetString("is-effective-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return hosRulesetAssignmentsListOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		NoAuth:         noAuth,
		Limit:          limit,
		Offset:         offset,
		Sort:           sort,
		Broker:         broker,
		HosDay:         hosDay,
		User:           user,
		Driver:         driver,
		EffectiveAtMin: effectiveAtMin,
		EffectiveAtMax: effectiveAtMax,
		IsEffectiveAt:  isEffectiveAt,
	}, nil
}

func buildHosRulesetAssignmentRows(resp jsonAPIResponse) []hosRulesetAssignmentRow {
	rows := make([]hosRulesetAssignmentRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildHosRulesetAssignmentRow(resource))
	}
	return rows
}

func buildHosRulesetAssignmentRow(resource jsonAPIResource) hosRulesetAssignmentRow {
	row := hosRulesetAssignmentRow{
		ID:          resource.ID,
		RuleSetID:   stringAttr(resource.Attributes, "rule-set-id"),
		Name:        stringAttr(resource.Attributes, "name"),
		EffectiveAt: formatDateTime(stringAttr(resource.Attributes, "effective-at")),
	}
	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		row.UserID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["hos-day"]; ok && rel.Data != nil {
		row.HosDayID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}
	return row
}

func buildHosRulesetAssignmentRowFromSingle(resp jsonAPISingleResponse) hosRulesetAssignmentRow {
	return buildHosRulesetAssignmentRow(resp.Data)
}

func renderHosRulesetAssignmentsTable(cmd *cobra.Command, rows []hosRulesetAssignmentRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No HOS ruleset assignments found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tRULE SET\tNAME\tEFFECTIVE AT\tUSER\tHOS DAY\tBROKER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.RuleSetID, 30),
			truncateString(row.Name, 30),
			row.EffectiveAt,
			row.UserID,
			row.HosDayID,
			row.BrokerID,
		)
	}
	return writer.Flush()
}
