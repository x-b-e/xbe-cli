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

type maintenanceRulesListOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	NoAuth         bool
	Limit          int
	Offset         int
	Me             bool
	BusinessUnitID string
	ActiveOnly     bool
	EquipmentID    string
	Sort           string
}

func newMaintenanceRulesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List maintenance requirement rules",
		Long: `List maintenance requirement rules with filtering and pagination.

Returns a list of rules that define when maintenance should be performed
on equipment.

Output Columns (table format):
  ID              Unique rule identifier
  RULE            Rule name
  CLASSIFICATION  Equipment classification the rule applies to
  EQUIPMENT       Specific equipment (or - if classification-level)
  BUSINESS_UNIT   Direct business unit owner (or - if none)
  SCOPE           Rule scope level:
                    Equipment: {name}  - targets specific equipment
                    {classification}   - targets equipment classification
                    BU: {name}         - targets specific business unit
                    Branch Level       - broker-wide rule
  ACTIVE          Whether the rule is active

Filtering:
  --me               Show rules for my business units
  --bu-id            Filter by business unit ID
  --equipment-id     Filter by equipment ID
  --active-only      Show only active rules

Rule Ownership:
  A rule belongs to a BU if ANY of these conditions is true:
  1. Rule has direct business_unit_id matching the BU
  2. Rule's equipment belongs to the BU
  3. Rule's classification matches BU's equipment AND rule has no specific equipment

  This means classification-level rules appear for ALL BUs with that equipment type.

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.
  Default: -created-at (newest first)`,
		Example: `  # List all rules
  xbe view maintenance rules list

  # List rules for my business units
  xbe view maintenance rules list --me

  # List only active rules
  xbe view maintenance rules list --active-only

  # Filter by business unit (complex ownership check)
  xbe view maintenance rules list --bu-id 123

  # Filter by equipment
  xbe view maintenance rules list --equipment-id 456

  # Combine filters
  xbe view maintenance rules list --me --active-only

  # Paginate results
  xbe view maintenance rules list --limit 50 --offset 100

  # Output as JSON
  xbe view maintenance rules list --json`,
		RunE: runMaintenanceRulesList,
	}
	initMaintenanceRulesListFlags(cmd)
	return cmd
}

func init() {
	maintenanceRulesCmd.AddCommand(newMaintenanceRulesListCmd())
}

func initMaintenanceRulesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Bool("me", false, "Show rules for my business units")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("bu-id", "", "Filter by business unit ID")
	cmd.Flags().String("equipment-id", "", "Filter by equipment ID")
	cmd.Flags().Bool("active-only", false, "Show only active rules")
	cmd.Flags().String("sort", "", "Sort order (default: -created-at)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaintenanceRulesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaintenanceRulesListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	// Determine if we need client-side BU filtering
	var buContext *BUEquipmentContext
	if opts.Me {
		if opts.BusinessUnitID != "" {
			return fmt.Errorf("cannot use both --me and --bu-id")
		}
		buIDs, err := getCurrentUserBusinessUnitIDs(cmd, client)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		ctx, err := getBUEquipmentContext(cmd, client, buIDs)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		buContext = ctx
	} else if opts.BusinessUnitID != "" {
		buIDs := []string{opts.BusinessUnitID}
		ctx, err := getBUEquipmentContext(cmd, client, buIDs)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		buContext = ctx
	}

	query := url.Values{}
	query.Set("include", "equipment,equipment-classification,business-unit,maintenance-requirement-sets")

	// When doing client-side filtering, we need to fetch more records
	// Use a larger page size and handle pagination manually
	if buContext != nil {
		// For client-side filtering, fetch a larger batch
		query.Set("page[limit]", "500")
	} else {
		if opts.Limit > 0 {
			query.Set("page[limit]", strconv.Itoa(opts.Limit))
		}
		if opts.Offset > 0 {
			query.Set("page[offset]", strconv.Itoa(opts.Offset))
		}
	}

	// Apply direct API filters (when not doing complex BU filtering)
	setFilterIfPresent(query, "filter[equipment]", opts.EquipmentID)
	if opts.ActiveOnly {
		query.Set("filter[is_active]", "true")
	}

	// Apply sort
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	} else {
		query.Set("sort", "-created-at")
	}

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

	// Apply client-side BU filtering if needed
	if buContext != nil {
		resp = filterRulesByBUContext(resp, buContext, opts.Limit, opts.Offset)
	}

	if opts.JSON {
		rows := buildRuleRows(resp)
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderRulesList(cmd, resp)
}

func parseMaintenanceRulesListOptions(cmd *cobra.Command) (maintenanceRulesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	me, _ := cmd.Flags().GetBool("me")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	businessUnitID, _ := cmd.Flags().GetString("bu-id")
	equipmentID, _ := cmd.Flags().GetString("equipment-id")
	activeOnly, _ := cmd.Flags().GetBool("active-only")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return maintenanceRulesListOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		NoAuth:         noAuth,
		Limit:          limit,
		Offset:         offset,
		Me:             me,
		BusinessUnitID: businessUnitID,
		EquipmentID:    equipmentID,
		ActiveOnly:     activeOnly,
		Sort:           sort,
	}, nil
}

// filterRulesByBUContext applies client-side filtering using the BU equipment context
func filterRulesByBUContext(resp jsonAPIResponse, ctx *BUEquipmentContext, limit, offset int) jsonAPIResponse {
	filtered := make([]jsonAPIResource, 0)
	for _, rule := range resp.Data {
		if canAccessRule(rule, ctx) {
			filtered = append(filtered, rule)
		}
	}

	// Apply offset and limit after filtering
	start := offset
	if start > len(filtered) {
		start = len(filtered)
	}
	end := len(filtered)
	if limit > 0 && start+limit < end {
		end = start + limit
	}

	return jsonAPIResponse{
		Data:     filtered[start:end],
		Included: resp.Included,
	}
}

func buildRuleRows(resp jsonAPIResponse) []ruleRow {
	// Build included map for lookups
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]ruleRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes

		// The API uses "rule" attribute for the rule name
		name := strings.TrimSpace(stringAttr(attrs, "rule"))
		if name == "" {
			name = strings.TrimSpace(stringAttr(attrs, "name"))
		}

		row := ruleRow{
			ID:              resource.ID,
			Name:            name,
			MaintenanceType: stringAttr(attrs, "maintenance-type"),
			IsActive:        boolAttr(attrs, "is-active"),
		}

		// Get equipment
		if rel, ok := resource.Relationships["equipment"]; ok && rel.Data != nil {
			row.EquipmentID = rel.Data.ID
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if inc, ok := included[key]; ok {
				row.Equipment = firstNonEmpty(
					stringAttr(inc.Attributes, "name"),
					stringAttr(inc.Attributes, "equipment-number"),
					rel.Data.ID,
				)
			}
		}

		// Get equipment classification
		if rel, ok := resource.Relationships["equipment-classification"]; ok && rel.Data != nil {
			row.EquipmentClassificationID = rel.Data.ID
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if inc, ok := included[key]; ok {
				row.EquipmentClassification = stringAttr(inc.Attributes, "name")
			}
		}

		// Get business unit
		if rel, ok := resource.Relationships["business-unit"]; ok && rel.Data != nil {
			row.BusinessUnitID = rel.Data.ID
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if inc, ok := included[key]; ok {
				row.BusinessUnit = firstNonEmpty(
					stringAttr(inc.Attributes, "company-name"),
					stringAttr(inc.Attributes, "name"),
					rel.Data.ID,
				)
			}
		}

		// Compute scope level and display
		row.ScopeLevel, row.Scope = getRuleScopeInfo(
			row.EquipmentID, row.Equipment,
			row.EquipmentClassificationID, row.EquipmentClassification,
			row.BusinessUnitID, row.BusinessUnit,
		)

		rows = append(rows, row)
	}
	return rows
}

func renderRulesList(cmd *cobra.Command, resp jsonAPIResponse) error {
	rows := buildRuleRows(resp)
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No rules found.")
		return nil
	}

	const ruleMax = 25
	const classMax = 18
	const equipMax = 15
	const buMax = 15
	const scopeMax = 18

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tRULE\tCLASSIFICATION\tEQUIPMENT\tBUSINESS_UNIT\tSCOPE\tACTIVE")
	for _, row := range rows {
		ruleName := row.Name
		if ruleName == "" {
			ruleName = row.MaintenanceType
		}
		if ruleName == "" {
			ruleName = "-"
		}
		active := "false"
		if row.IsActive {
			active = "true"
		}
		classification := row.EquipmentClassification
		if classification == "" {
			classification = "-"
		}
		equipment := row.Equipment
		if equipment == "" {
			equipment = "-"
		}
		bu := row.BusinessUnit
		if bu == "" {
			bu = "-"
		}
		scope := row.Scope
		if scope == "" {
			scope = "-"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(ruleName, ruleMax),
			truncateString(classification, classMax),
			truncateString(equipment, equipMax),
			truncateString(bu, buMax),
			truncateString(scope, scopeMax),
			active,
		)
	}
	return writer.Flush()
}
