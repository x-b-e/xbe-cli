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

type driverAssignmentRulesListOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	NoAuth    bool
	Limit     int
	Offset    int
	Sort      string
	LevelType string
	LevelID   string
	Broker    string
}

type driverAssignmentRuleRow struct {
	ID        string `json:"id"`
	Rule      string `json:"rule,omitempty"`
	IsActive  bool   `json:"is_active"`
	LevelType string `json:"level_type,omitempty"`
	LevelID   string `json:"level_id,omitempty"`
	BrokerID  string `json:"broker_id,omitempty"`
}

func newDriverAssignmentRulesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List driver assignment rules",
		Long: `List driver assignment rules.

Output Columns:
  ID      Driver assignment rule identifier
  ACTIVE  Whether the rule is active
  LEVEL   Level type and ID
  BROKER  Broker ID
  RULE    Rule text

Filters:
  --level-type  Filter by level type (Broker, JobScheduleShift, Project, JobProductionPlan, MaterialSupplier, MaterialSite, MaterialType, Trucker, JobSite)
  --level-id    Filter by level ID (used with --level-type)
  --broker      Filter by broker ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List driver assignment rules
  xbe view driver-assignment-rules list

  # Filter by broker
  xbe view driver-assignment-rules list --broker 123

  # Filter by level
  xbe view driver-assignment-rules list --level-type Broker --level-id 456

  # Output as JSON
  xbe view driver-assignment-rules list --json`,
		Args: cobra.NoArgs,
		RunE: runDriverAssignmentRulesList,
	}
	initDriverAssignmentRulesListFlags(cmd)
	return cmd
}

func init() {
	driverAssignmentRulesCmd.AddCommand(newDriverAssignmentRulesListCmd())
}

func initDriverAssignmentRulesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("level-type", "", "Filter by level type (Broker, JobScheduleShift, Project, JobProductionPlan, MaterialSupplier, MaterialSite, MaterialType, Trucker, JobSite)")
	cmd.Flags().String("level-id", "", "Filter by level ID (used with --level-type)")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDriverAssignmentRulesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseDriverAssignmentRulesListOptions(cmd)
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
	query.Set("fields[driver-assignment-rules]", "rule,is-active,level,broker")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	if opts.LevelType != "" && opts.LevelID != "" {
		levelType := normalizeDriverAssignmentRuleLevelFilter(opts.LevelType)
		query.Set("filter[level]", levelType+"|"+opts.LevelID)
	}
	setFilterIfPresent(query, "filter[broker]", opts.Broker)

	body, _, err := client.Get(cmd.Context(), "/v1/driver-assignment-rules", query)
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

	rows := buildDriverAssignmentRuleRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderDriverAssignmentRulesTable(cmd, rows)
}

func parseDriverAssignmentRulesListOptions(cmd *cobra.Command) (driverAssignmentRulesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	levelType, _ := cmd.Flags().GetString("level-type")
	levelID, _ := cmd.Flags().GetString("level-id")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return driverAssignmentRulesListOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		NoAuth:    noAuth,
		Limit:     limit,
		Offset:    offset,
		Sort:      sort,
		LevelType: levelType,
		LevelID:   levelID,
		Broker:    broker,
	}, nil
}

func buildDriverAssignmentRuleRows(resp jsonAPIResponse) []driverAssignmentRuleRow {
	rows := make([]driverAssignmentRuleRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildDriverAssignmentRuleRow(resource))
	}
	return rows
}

func buildDriverAssignmentRuleRow(resource jsonAPIResource) driverAssignmentRuleRow {
	row := driverAssignmentRuleRow{
		ID:       resource.ID,
		Rule:     stringAttr(resource.Attributes, "rule"),
		IsActive: boolAttr(resource.Attributes, "is-active"),
	}
	if rel, ok := resource.Relationships["level"]; ok && rel.Data != nil {
		row.LevelType = rel.Data.Type
		row.LevelID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}
	return row
}

func buildDriverAssignmentRuleRowFromSingle(resp jsonAPISingleResponse) driverAssignmentRuleRow {
	return buildDriverAssignmentRuleRow(resp.Data)
}

func renderDriverAssignmentRulesTable(cmd *cobra.Command, rows []driverAssignmentRuleRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No driver assignment rules found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tACTIVE\tLEVEL\tBROKER\tRULE")
	for _, row := range rows {
		level := ""
		if row.LevelType != "" {
			level = row.LevelType
			if row.LevelID != "" {
				level += "/" + row.LevelID
			}
		} else if row.LevelID != "" {
			level = row.LevelID
		}
		fmt.Fprintf(writer, "%s\t%t\t%s\t%s\t%s\n",
			row.ID,
			row.IsActive,
			truncateString(level, 30),
			row.BrokerID,
			truncateString(row.Rule, 50),
		)
	}
	return writer.Flush()
}
