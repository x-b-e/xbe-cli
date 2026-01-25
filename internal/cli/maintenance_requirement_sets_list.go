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

type maintenanceRequirementSetsListOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	NoAuth                  bool
	Limit                   int
	Offset                  int
	Sort                    string
	EquipmentClassification string
	Broker                  string
	EquipmentBusinessUnit   string
	Equipment               string
	MaintenanceType         string
	Status                  string
	IsTemplate              string
	IsArchived              string
	CompletedAtMin          string
	CompletedAtMax          string
	Q                       string
}

type maintenanceRequirementSetRow struct {
	ID                            string `json:"id"`
	MaintenanceType               string `json:"maintenance_type,omitempty"`
	Status                        string `json:"status,omitempty"`
	IsTemplate                    bool   `json:"is_template"`
	TemplateName                  string `json:"template_name,omitempty"`
	IsArchived                    bool   `json:"is_archived"`
	CompletedAt                   string `json:"completed_at,omitempty"`
	BrokerID                      string `json:"broker_id,omitempty"`
	BrokerName                    string `json:"broker_name,omitempty"`
	EquipmentClassificationID     string `json:"equipment_classification_id,omitempty"`
	EquipmentClassification       string `json:"equipment_classification,omitempty"`
	EquipmentClassificationAbbrev string `json:"equipment_classification_abbreviation,omitempty"`
	WorkOrderID                   string `json:"work_order_id,omitempty"`
}

func newMaintenanceRequirementSetsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List maintenance requirement sets",
		Long: `List maintenance requirement sets.

Output Columns:
  ID          Maintenance requirement set identifier
  TYPE        Maintenance type (inspection/maintenance)
  STATUS      Current status
  TEMPLATE    Template name or "no"
  ARCHIVED    Archived status
  COMPLETED   Completed timestamp
  BROKER      Broker
  EQUIP CLASS Equipment classification
  WORK ORDER  Work order ID (if linked)

Filters:
  --equipment-classification  Filter by equipment classification ID
  --broker                    Filter by broker ID
  --equipment-business-unit   Filter by equipment business unit ID
  --equipment                 Filter by equipment ID
  --maintenance-type          Filter by maintenance type
  --status                    Filter by status
  --is-template               Filter by template status (true/false)
  --is-archived               Filter by archived status (true/false)
  --completed-at-min          Filter by completion on/after (ISO 8601)
  --completed-at-max          Filter by completion on/before (ISO 8601)
  --q                         Search by template name

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List maintenance requirement sets
  xbe view maintenance-requirement-sets list

  # Filter by broker and status
  xbe view maintenance-requirement-sets list --broker 123 --status ready_for_work

  # Filter by template sets
  xbe view maintenance-requirement-sets list --is-template true

  # JSON output
  xbe view maintenance-requirement-sets list --json`,
		Args: cobra.NoArgs,
		RunE: runMaintenanceRequirementSetsList,
	}
	initMaintenanceRequirementSetsListFlags(cmd)
	return cmd
}

func init() {
	maintenanceRequirementSetsCmd.AddCommand(newMaintenanceRequirementSetsListCmd())
}

func initMaintenanceRequirementSetsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("equipment-classification", "", "Filter by equipment classification ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("equipment-business-unit", "", "Filter by equipment business unit ID")
	cmd.Flags().String("equipment", "", "Filter by equipment ID")
	cmd.Flags().String("maintenance-type", "", "Filter by maintenance type")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("is-template", "", "Filter by template status (true/false)")
	cmd.Flags().String("is-archived", "", "Filter by archived status (true/false)")
	cmd.Flags().String("completed-at-min", "", "Filter by completion on/after (ISO 8601)")
	cmd.Flags().String("completed-at-max", "", "Filter by completion on/before (ISO 8601)")
	cmd.Flags().String("q", "", "Search by template name")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaintenanceRequirementSetsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaintenanceRequirementSetsListOptions(cmd)
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
	query.Set("fields[maintenance-requirement-sets]", "maintenance-type,status,is-template,template-name,is-archived,completed-at,broker,equipment-classification,work-order")
	query.Set("include", "broker,equipment-classification")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[equipment-classifications]", "name,abbreviation")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[equipment_classification]", opts.EquipmentClassification)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[equipment_business_unit]", opts.EquipmentBusinessUnit)
	setFilterIfPresent(query, "filter[equipment]", opts.Equipment)
	setFilterIfPresent(query, "filter[maintenance-type]", opts.MaintenanceType)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[is-template]", opts.IsTemplate)
	setFilterIfPresent(query, "filter[is-archived]", opts.IsArchived)
	setFilterIfPresent(query, "filter[completed-at-min]", opts.CompletedAtMin)
	setFilterIfPresent(query, "filter[completed-at-max]", opts.CompletedAtMax)
	setFilterIfPresent(query, "filter[q]", opts.Q)

	body, _, err := client.Get(cmd.Context(), "/v1/maintenance-requirement-sets", query)
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

	rows := buildMaintenanceRequirementSetRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaintenanceRequirementSetsTable(cmd, rows)
}

func parseMaintenanceRequirementSetsListOptions(cmd *cobra.Command) (maintenanceRequirementSetsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	equipmentClassification, _ := cmd.Flags().GetString("equipment-classification")
	broker, _ := cmd.Flags().GetString("broker")
	equipmentBusinessUnit, _ := cmd.Flags().GetString("equipment-business-unit")
	equipment, _ := cmd.Flags().GetString("equipment")
	maintenanceType, _ := cmd.Flags().GetString("maintenance-type")
	status, _ := cmd.Flags().GetString("status")
	isTemplate, _ := cmd.Flags().GetString("is-template")
	isArchived, _ := cmd.Flags().GetString("is-archived")
	completedAtMin, _ := cmd.Flags().GetString("completed-at-min")
	completedAtMax, _ := cmd.Flags().GetString("completed-at-max")
	q, _ := cmd.Flags().GetString("q")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return maintenanceRequirementSetsListOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		NoAuth:                  noAuth,
		Limit:                   limit,
		Offset:                  offset,
		Sort:                    sort,
		EquipmentClassification: equipmentClassification,
		Broker:                  broker,
		EquipmentBusinessUnit:   equipmentBusinessUnit,
		Equipment:               equipment,
		MaintenanceType:         maintenanceType,
		Status:                  status,
		IsTemplate:              isTemplate,
		IsArchived:              isArchived,
		CompletedAtMin:          completedAtMin,
		CompletedAtMax:          completedAtMax,
		Q:                       q,
	}, nil
}

func buildMaintenanceRequirementSetRows(resp jsonAPIResponse) []maintenanceRequirementSetRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]maintenanceRequirementSetRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := maintenanceRequirementSetRow{
			ID:              resource.ID,
			MaintenanceType: stringAttr(attrs, "maintenance-type"),
			Status:          stringAttr(attrs, "status"),
			IsTemplate:      boolAttr(attrs, "is-template"),
			TemplateName:    strings.TrimSpace(stringAttr(attrs, "template-name")),
			IsArchived:      boolAttr(attrs, "is-archived"),
			CompletedAt:     formatDateTime(stringAttr(attrs, "completed-at")),
		}

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
			if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.BrokerName = stringAttr(broker.Attributes, "company-name")
			}
		}
		if rel, ok := resource.Relationships["equipment-classification"]; ok && rel.Data != nil {
			row.EquipmentClassificationID = rel.Data.ID
			if ec, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.EquipmentClassification = stringAttr(ec.Attributes, "name")
				row.EquipmentClassificationAbbrev = stringAttr(ec.Attributes, "abbreviation")
			}
		}
		if rel, ok := resource.Relationships["work-order"]; ok && rel.Data != nil {
			row.WorkOrderID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func buildMaintenanceRequirementSetRowFromSingle(resp jsonAPISingleResponse) maintenanceRequirementSetRow {
	rows := buildMaintenanceRequirementSetRows(jsonAPIResponse{
		Data:     []jsonAPIResource{resp.Data},
		Included: resp.Included,
	})
	if len(rows) == 0 {
		return maintenanceRequirementSetRow{}
	}
	return rows[0]
}

func renderMaintenanceRequirementSetsTable(cmd *cobra.Command, rows []maintenanceRequirementSetRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No maintenance requirement sets found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTYPE\tSTATUS\tTEMPLATE\tARCHIVED\tCOMPLETED\tBROKER\tEQUIP CLASS\tWORK ORDER")
	for _, row := range rows {
		templateLabel := "no"
		if row.IsTemplate {
			templateLabel = "yes"
			if row.TemplateName != "" {
				templateLabel = row.TemplateName
			}
		}
		archived := "no"
		if row.IsArchived {
			archived = "yes"
		}
		brokerLabel := firstNonEmpty(row.BrokerName, row.BrokerID)
		equipmentLabel := formatEquipmentClassificationLabel(row.EquipmentClassification, row.EquipmentClassificationAbbrev)
		equipmentLabel = firstNonEmpty(equipmentLabel, row.EquipmentClassificationID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.MaintenanceType,
			row.Status,
			truncateString(templateLabel, 30),
			archived,
			row.CompletedAt,
			truncateString(brokerLabel, 25),
			truncateString(equipmentLabel, 25),
			row.WorkOrderID,
		)
	}
	return writer.Flush()
}

func formatEquipmentClassificationLabel(name, abbreviation string) string {
	name = strings.TrimSpace(name)
	abbreviation = strings.TrimSpace(abbreviation)
	if name != "" && abbreviation != "" {
		return fmt.Sprintf("%s (%s)", name, abbreviation)
	}
	return firstNonEmpty(name, abbreviation)
}
