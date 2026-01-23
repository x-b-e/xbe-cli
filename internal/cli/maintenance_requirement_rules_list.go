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

type maintenanceRequirementRulesListOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	NoAuth                  bool
	Limit                   int
	Offset                  int
	Sort                    string
	Broker                  string
	Equipment               string
	EquipmentClassification string
	BusinessUnit            string
	IsActive                string
}

type maintenanceRequirementRuleRow struct {
	ID                        string `json:"id"`
	Rule                      string `json:"rule,omitempty"`
	IsActive                  bool   `json:"is_active"`
	BrokerID                  string `json:"broker_id,omitempty"`
	EquipmentID               string `json:"equipment_id,omitempty"`
	EquipmentClassificationID string `json:"equipment_classification_id,omitempty"`
	BusinessUnitID            string `json:"business_unit_id,omitempty"`
}

func newMaintenanceRequirementRulesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List maintenance requirement rules",
		Long: `List maintenance requirement rules.

Output Columns:
  ID             Maintenance requirement rule identifier
  ACTIVE         Whether the rule is active
  BROKER         Broker ID
  EQUIPMENT      Equipment ID
  EQUIP CLASS    Equipment classification ID
  BUSINESS UNIT  Business unit ID
  RULE           Rule text

Filters:
  --broker                   Filter by broker ID
  --equipment                Filter by equipment ID
  --equipment-classification Filter by equipment classification ID
  --business-unit            Filter by business unit ID
  --is-active                Filter by active status (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List maintenance requirement rules
  xbe view maintenance-requirement-rules list

  # Filter by broker
  xbe view maintenance-requirement-rules list --broker 123

  # Filter by equipment
  xbe view maintenance-requirement-rules list --equipment 234

  # Filter by equipment classification
  xbe view maintenance-requirement-rules list --equipment-classification 456

  # Filter by business unit
  xbe view maintenance-requirement-rules list --business-unit 789

  # Filter by active status
  xbe view maintenance-requirement-rules list --is-active false

  # Output as JSON
  xbe view maintenance-requirement-rules list --json`,
		Args: cobra.NoArgs,
		RunE: runMaintenanceRequirementRulesList,
	}
	initMaintenanceRequirementRulesListFlags(cmd)
	return cmd
}

func init() {
	maintenanceRequirementRulesCmd.AddCommand(newMaintenanceRequirementRulesListCmd())
}

func initMaintenanceRequirementRulesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("equipment", "", "Filter by equipment ID")
	cmd.Flags().String("equipment-classification", "", "Filter by equipment classification ID")
	cmd.Flags().String("business-unit", "", "Filter by business unit ID")
	cmd.Flags().String("is-active", "", "Filter by active status (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaintenanceRequirementRulesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaintenanceRequirementRulesListOptions(cmd)
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
	query.Set("fields[maintenance-requirement-rules]", "rule,is-active,broker,equipment,equipment-classification,business-unit")

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
	setFilterIfPresent(query, "filter[equipment]", opts.Equipment)
	setFilterIfPresent(query, "filter[equipment_classification]", opts.EquipmentClassification)
	setFilterIfPresent(query, "filter[business_unit]", opts.BusinessUnit)
	setFilterIfPresent(query, "filter[is_active]", opts.IsActive)

	body, _, err := client.Get(cmd.Context(), "/v1/maintenance-requirement-rules", query)
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

	rows := buildMaintenanceRequirementRuleRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaintenanceRequirementRulesTable(cmd, rows)
}

func parseMaintenanceRequirementRulesListOptions(cmd *cobra.Command) (maintenanceRequirementRulesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	equipment, _ := cmd.Flags().GetString("equipment")
	equipmentClassification, _ := cmd.Flags().GetString("equipment-classification")
	businessUnit, _ := cmd.Flags().GetString("business-unit")
	isActive, _ := cmd.Flags().GetString("is-active")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return maintenanceRequirementRulesListOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		NoAuth:                  noAuth,
		Limit:                   limit,
		Offset:                  offset,
		Sort:                    sort,
		Broker:                  broker,
		Equipment:               equipment,
		EquipmentClassification: equipmentClassification,
		BusinessUnit:            businessUnit,
		IsActive:                isActive,
	}, nil
}

func buildMaintenanceRequirementRuleRows(resp jsonAPIResponse) []maintenanceRequirementRuleRow {
	rows := make([]maintenanceRequirementRuleRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildMaintenanceRequirementRuleRow(resource))
	}
	return rows
}

func buildMaintenanceRequirementRuleRow(resource jsonAPIResource) maintenanceRequirementRuleRow {
	row := maintenanceRequirementRuleRow{
		ID:       resource.ID,
		Rule:     stringAttr(resource.Attributes, "rule"),
		IsActive: boolAttr(resource.Attributes, "is-active"),
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["equipment"]; ok && rel.Data != nil {
		row.EquipmentID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["equipment-classification"]; ok && rel.Data != nil {
		row.EquipmentClassificationID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["business-unit"]; ok && rel.Data != nil {
		row.BusinessUnitID = rel.Data.ID
	}

	return row
}

func buildMaintenanceRequirementRuleRowFromSingle(resp jsonAPISingleResponse) maintenanceRequirementRuleRow {
	return buildMaintenanceRequirementRuleRow(resp.Data)
}

func renderMaintenanceRequirementRulesTable(cmd *cobra.Command, rows []maintenanceRequirementRuleRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No maintenance requirement rules found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tACTIVE\tBROKER\tEQUIPMENT\tEQUIP CLASS\tBUSINESS UNIT\tRULE")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%t\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.IsActive,
			row.BrokerID,
			row.EquipmentID,
			row.EquipmentClassificationID,
			row.BusinessUnitID,
			truncateString(row.Rule, 50),
		)
	}
	return writer.Flush()
}
