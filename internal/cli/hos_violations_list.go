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

type hosViolationsListOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	NoAuth     bool
	Limit      int
	Offset     int
	Sort       string
	Broker     string
	HosDay     string
	User       string
	Driver     string
	StartAtMin string
	StartAtMax string
	EndAtMin   string
	EndAtMax   string
}

type hosViolationRow struct {
	ID                string `json:"id"`
	RegulationSetCode string `json:"regulation_set_code,omitempty"`
	ViolationType     string `json:"violation_type,omitempty"`
	StartAt           string `json:"start_at,omitempty"`
	EndAt             string `json:"end_at,omitempty"`
	RuleID            string `json:"rule_id,omitempty"`
	RuleName          string `json:"rule_name,omitempty"`
	BrokerID          string `json:"broker_id,omitempty"`
	HosDayID          string `json:"hos_day_id,omitempty"`
	UserID            string `json:"user_id,omitempty"`
}

func newHosViolationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List HOS violations",
		Long: `List HOS violations with filtering and pagination.

HOS violations capture hours-of-service rule breaches for a driver.

Output Columns:
  ID       Violation identifier
  DRIVER   Driver user ID
  HOS DAY  HOS day ID
  TYPE     Violation type
  START    Violation start timestamp
  END      Violation end timestamp
  RULE     Rule name or ID
  REG SET  Regulation set code

Filters:
  --broker        Filter by broker ID
  --hos-day       Filter by HOS day ID
  --user          Filter by user ID
  --driver        Filter by driver user ID (alias for user)
  --start-at-min  Filter by start-at on/after (ISO 8601)
  --start-at-max  Filter by start-at on/before (ISO 8601)
  --end-at-min    Filter by end-at on/after (ISO 8601)
  --end-at-max    Filter by end-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List HOS violations
  xbe view hos-violations list

  # Filter by driver
  xbe view hos-violations list --driver 123

  # Filter by HOS day
  xbe view hos-violations list --hos-day 456

  # Filter by time window
  xbe view hos-violations list --start-at-min 2025-01-01T00:00:00Z --end-at-max 2025-01-02T00:00:00Z

  # Output as JSON
  xbe view hos-violations list --json`,
		Args: cobra.NoArgs,
		RunE: runHosViolationsList,
	}
	initHosViolationsListFlags(cmd)
	return cmd
}

func init() {
	hosViolationsCmd.AddCommand(newHosViolationsListCmd())
}

func initHosViolationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("hos-day", "", "Filter by HOS day ID")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("driver", "", "Filter by driver user ID")
	cmd.Flags().String("start-at-min", "", "Filter by start-at on/after (ISO 8601)")
	cmd.Flags().String("start-at-max", "", "Filter by start-at on/before (ISO 8601)")
	cmd.Flags().String("end-at-min", "", "Filter by end-at on/after (ISO 8601)")
	cmd.Flags().String("end-at-max", "", "Filter by end-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runHosViolationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseHosViolationsListOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[hos-violations]", "regulation-set-code,violation-type,start-at,end-at,rule-id,rule-name")

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
	setFilterIfPresent(query, "filter[hos_day]", opts.HosDay)
	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[driver]", opts.Driver)
	setFilterIfPresent(query, "filter[start-at-min]", opts.StartAtMin)
	setFilterIfPresent(query, "filter[start-at-max]", opts.StartAtMax)
	setFilterIfPresent(query, "filter[end-at-min]", opts.EndAtMin)
	setFilterIfPresent(query, "filter[end-at-max]", opts.EndAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/hos-violations", query)
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

	rows := buildHosViolationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderHosViolationsTable(cmd, rows)
}

func parseHosViolationsListOptions(cmd *cobra.Command) (hosViolationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	hosDay, _ := cmd.Flags().GetString("hos-day")
	user, _ := cmd.Flags().GetString("user")
	driver, _ := cmd.Flags().GetString("driver")
	startAtMin, _ := cmd.Flags().GetString("start-at-min")
	startAtMax, _ := cmd.Flags().GetString("start-at-max")
	endAtMin, _ := cmd.Flags().GetString("end-at-min")
	endAtMax, _ := cmd.Flags().GetString("end-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return hosViolationsListOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		NoAuth:     noAuth,
		Limit:      limit,
		Offset:     offset,
		Sort:       sort,
		Broker:     broker,
		HosDay:     hosDay,
		User:       user,
		Driver:     driver,
		StartAtMin: startAtMin,
		StartAtMax: startAtMax,
		EndAtMin:   endAtMin,
		EndAtMax:   endAtMax,
	}, nil
}

func buildHosViolationRows(resp jsonAPIResponse) []hosViolationRow {
	rows := make([]hosViolationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := hosViolationRow{
			ID:                resource.ID,
			RegulationSetCode: stringAttr(attrs, "regulation-set-code"),
			ViolationType:     stringAttr(attrs, "violation-type"),
			StartAt:           formatDateTime(stringAttr(attrs, "start-at")),
			EndAt:             formatDateTime(stringAttr(attrs, "end-at")),
			RuleID:            stringAttr(attrs, "rule-id"),
			RuleName:          stringAttr(attrs, "rule-name"),
		}

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["hos-day"]; ok && rel.Data != nil {
			row.HosDayID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
			row.UserID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderHosViolationsTable(cmd *cobra.Command, rows []hosViolationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No HOS violations found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tDRIVER\tHOS DAY\tTYPE\tSTART\tEND\tRULE\tREG SET")
	for _, row := range rows {
		rule := firstNonEmpty(row.RuleName, row.RuleID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.UserID,
			row.HosDayID,
			row.ViolationType,
			row.StartAt,
			row.EndAt,
			truncateString(rule, 30),
			row.RegulationSetCode,
		)
	}
	return writer.Flush()
}
