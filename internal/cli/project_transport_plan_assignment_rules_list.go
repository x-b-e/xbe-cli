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

type projectTransportPlanAssignmentRulesListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	AssetType    string
	Level        string
	LevelType    string
	LevelID      string
	NotLevelType string
	Broker       string
	IsActive     string
}

type projectTransportPlanAssignmentRuleRow struct {
	ID        string `json:"id"`
	Rule      string `json:"rule,omitempty"`
	AssetType string `json:"asset_type,omitempty"`
	IsActive  bool   `json:"is_active"`
	LevelType string `json:"level_type,omitempty"`
	LevelID   string `json:"level_id,omitempty"`
	BrokerID  string `json:"broker_id,omitempty"`
	Broker    string `json:"broker,omitempty"`
}

func newProjectTransportPlanAssignmentRulesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project transport plan assignment rules",
		Long: `List project transport plan assignment rules with filtering and pagination.

Project transport plan assignment rules define broker-level rules for assigning
project transport plan drivers, tractors, and trailers.

Output Columns:
  ID         Assignment rule identifier
  RULE       Rule text (truncated)
  ASSET TYPE Asset type (driver/tractor/trailer)
  ACTIVE     Whether the rule is active
  LEVEL      Level type and ID
  BROKER     Broker name

Filters:
  --asset-type     Filter by asset type (driver/tractor/trailer)
  --level          Filter by level (format: Type|ID, e.g., Broker|123)
  --level-type     Filter by level type (e.g., Broker)
  --level-id       Filter by level ID (use with --level-type)
  --not-level-type Exclude a level type (e.g., Customer)
  --broker         Filter by broker ID
  --is-active      Filter by active status (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List assignment rules
  xbe view project-transport-plan-assignment-rules list

  # Filter by asset type
  xbe view project-transport-plan-assignment-rules list --asset-type driver

  # Filter by broker level
  xbe view project-transport-plan-assignment-rules list --level "Broker|123"

  # Output as JSON
  xbe view project-transport-plan-assignment-rules list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectTransportPlanAssignmentRulesList,
	}
	initProjectTransportPlanAssignmentRulesListFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanAssignmentRulesCmd.AddCommand(newProjectTransportPlanAssignmentRulesListCmd())
}

func initProjectTransportPlanAssignmentRulesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("asset-type", "", "Filter by asset type (driver/tractor/trailer)")
	cmd.Flags().String("level", "", "Filter by level (format: Type|ID)")
	cmd.Flags().String("level-type", "", "Filter by level type")
	cmd.Flags().String("level-id", "", "Filter by level ID (use with --level-type)")
	cmd.Flags().String("not-level-type", "", "Exclude a level type")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("is-active", "", "Filter by active status (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanAssignmentRulesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTransportPlanAssignmentRulesListOptions(cmd)
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
	query.Set("fields[project-transport-plan-assignment-rules]", "rule,asset-type,is-active,level,broker")
	query.Set("fields[brokers]", "company-name")
	query.Set("include", "broker")

	levelFilter := normalizePolymorphicFilterValue(opts.Level)
	if levelFilter == "" && opts.LevelID != "" {
		if opts.LevelType == "" {
			return fmt.Errorf("--level-id requires --level-type or --level")
		}
		levelFilter = normalizeResourceTypeForFilter(opts.LevelType) + "|" + strings.TrimSpace(opts.LevelID)
	}

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[asset-type]", opts.AssetType)
	setFilterIfPresent(query, "filter[level]", levelFilter)
	if levelFilter == "" {
		setFilterIfPresent(query, "filter[level-type]", normalizeResourceTypeForFilter(opts.LevelType))
	}
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[is-active]", opts.IsActive)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-assignment-rules", query)
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

	rows := buildProjectTransportPlanAssignmentRuleRows(resp)
	if opts.NotLevelType != "" {
		notLevelType := normalizeResourceTypeForFilter(opts.NotLevelType)
		if notLevelType != "" {
			filtered := rows[:0]
			for _, row := range rows {
				if normalizeResourceTypeForFilter(row.LevelType) == notLevelType {
					continue
				}
				filtered = append(filtered, row)
			}
			rows = filtered
		}
	}
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTransportPlanAssignmentRulesTable(cmd, rows)
}

func parseProjectTransportPlanAssignmentRulesListOptions(cmd *cobra.Command) (projectTransportPlanAssignmentRulesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	assetType, _ := cmd.Flags().GetString("asset-type")
	level, _ := cmd.Flags().GetString("level")
	levelType, _ := cmd.Flags().GetString("level-type")
	levelID, _ := cmd.Flags().GetString("level-id")
	notLevelType, _ := cmd.Flags().GetString("not-level-type")
	broker, _ := cmd.Flags().GetString("broker")
	isActive, _ := cmd.Flags().GetString("is-active")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanAssignmentRulesListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		AssetType:    assetType,
		Level:        level,
		LevelType:    levelType,
		LevelID:      levelID,
		NotLevelType: notLevelType,
		Broker:       broker,
		IsActive:     isActive,
	}, nil
}

func buildProjectTransportPlanAssignmentRuleRows(resp jsonAPIResponse) []projectTransportPlanAssignmentRuleRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]projectTransportPlanAssignmentRuleRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := projectTransportPlanAssignmentRuleRow{
			ID:        resource.ID,
			Rule:      stringAttr(resource.Attributes, "rule"),
			AssetType: stringAttr(resource.Attributes, "asset-type"),
			IsActive:  boolAttr(resource.Attributes, "is-active"),
		}

		if rel, ok := resource.Relationships["level"]; ok && rel.Data != nil {
			row.LevelType = rel.Data.Type
			row.LevelID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
			if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.Broker = stringAttr(broker.Attributes, "company-name")
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func buildProjectTransportPlanAssignmentRuleRowFromSingle(resp jsonAPISingleResponse) projectTransportPlanAssignmentRuleRow {
	resource := resp.Data
	row := projectTransportPlanAssignmentRuleRow{
		ID:        resource.ID,
		Rule:      stringAttr(resource.Attributes, "rule"),
		AssetType: stringAttr(resource.Attributes, "asset-type"),
		IsActive:  boolAttr(resource.Attributes, "is-active"),
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

func renderProjectTransportPlanAssignmentRulesTable(cmd *cobra.Command, rows []projectTransportPlanAssignmentRuleRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project transport plan assignment rules found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tRULE\tASSET TYPE\tACTIVE\tLEVEL\tBROKER")
	for _, row := range rows {
		active := "no"
		if row.IsActive {
			active = "yes"
		}
		level := ""
		if row.LevelType != "" && row.LevelID != "" {
			level = row.LevelType + "/" + row.LevelID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Rule, 30),
			truncateString(row.AssetType, 10),
			active,
			truncateString(level, 30),
			truncateString(row.Broker, 25),
		)
	}
	return writer.Flush()
}

func normalizePolymorphicFilterValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	parts := strings.SplitN(value, "|", 2)
	if len(parts) != 2 {
		return value
	}
	typePart := normalizeResourceTypeForFilter(parts[0])
	if typePart == "" {
		typePart = strings.TrimSpace(parts[0])
	}
	idPart := strings.TrimSpace(parts[1])
	if idPart == "" {
		return value
	}
	return typePart + "|" + idPart
}
