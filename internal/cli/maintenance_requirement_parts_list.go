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

type maintenanceRequirementPartsListOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	NoAuth                  bool
	Limit                   int
	Offset                  int
	Sort                    string
	Broker                  string
	EquipmentClassification string
	Make                    string
	Model                   string
	Year                    string
	IsTemplate              string
	MaintenanceRequirements string
}

type maintenanceRequirementPartRow struct {
	ID                        string `json:"id"`
	PartNumber                string `json:"part_number,omitempty"`
	Name                      string `json:"name,omitempty"`
	Make                      string `json:"make,omitempty"`
	Model                     string `json:"model,omitempty"`
	Year                      string `json:"year,omitempty"`
	IsTemplate                bool   `json:"is_template"`
	BrokerID                  string `json:"broker_id,omitempty"`
	EquipmentClassificationID string `json:"equipment_classification_id,omitempty"`
}

func newMaintenanceRequirementPartsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List maintenance requirement parts",
		Long: `List maintenance requirement parts.

Maintenance requirement parts represent parts associated with maintenance requirements
or reusable template parts tied to a broker.

Output Columns:
  ID           Maintenance requirement part identifier
  NAME         Part name
  PART NUMBER  Part number
  MAKE         Part make
  MODEL        Part model
  YEAR         Part year
  TEMPLATE     Whether the part is a template
  BROKER       Broker ID (template parts)
  EQUIP CLASS  Equipment classification ID

Filters:
  --broker                   Filter by broker ID
  --equipment-classification Filter by equipment classification ID
  --make                     Filter by make
  --model                    Filter by model
  --year                     Filter by year
  --is-template              Filter by template flag (true/false)
  --maintenance-requirements Filter by maintenance requirement ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List maintenance requirement parts
  xbe view maintenance-requirement-parts list

  # Filter by broker
  xbe view maintenance-requirement-parts list --broker 123

  # Filter by equipment classification
  xbe view maintenance-requirement-parts list --equipment-classification 456

  # Filter by template parts
  xbe view maintenance-requirement-parts list --is-template true

  # Filter by maintenance requirement
  xbe view maintenance-requirement-parts list --maintenance-requirements 789

  # Output as JSON
  xbe view maintenance-requirement-parts list --json`,
		Args: cobra.NoArgs,
		RunE: runMaintenanceRequirementPartsList,
	}
	initMaintenanceRequirementPartsListFlags(cmd)
	return cmd
}

func init() {
	maintenanceRequirementPartsCmd.AddCommand(newMaintenanceRequirementPartsListCmd())
}

func initMaintenanceRequirementPartsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("equipment-classification", "", "Filter by equipment classification ID")
	cmd.Flags().String("make", "", "Filter by make")
	cmd.Flags().String("model", "", "Filter by model")
	cmd.Flags().String("year", "", "Filter by year")
	cmd.Flags().String("is-template", "", "Filter by template flag (true/false)")
	cmd.Flags().String("maintenance-requirements", "", "Filter by maintenance requirement ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaintenanceRequirementPartsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaintenanceRequirementPartsListOptions(cmd)
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
	query.Set("fields[maintenance-requirement-parts]", "part-number,name,is-template,make,model,year,broker,equipment-classification")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[equipment_classification]", opts.EquipmentClassification)
	setFilterIfPresent(query, "filter[make]", opts.Make)
	setFilterIfPresent(query, "filter[model]", opts.Model)
	setFilterIfPresent(query, "filter[year]", opts.Year)
	setFilterIfPresent(query, "filter[is_template]", opts.IsTemplate)
	setFilterIfPresent(query, "filter[maintenance_requirements]", opts.MaintenanceRequirements)

	body, _, err := client.Get(cmd.Context(), "/v1/maintenance-requirement-parts", query)
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

	rows := buildMaintenanceRequirementPartRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaintenanceRequirementPartsTable(cmd, rows)
}

func parseMaintenanceRequirementPartsListOptions(cmd *cobra.Command) (maintenanceRequirementPartsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	equipmentClassification, _ := cmd.Flags().GetString("equipment-classification")
	make, _ := cmd.Flags().GetString("make")
	model, _ := cmd.Flags().GetString("model")
	year, _ := cmd.Flags().GetString("year")
	isTemplate, _ := cmd.Flags().GetString("is-template")
	maintenanceRequirements, _ := cmd.Flags().GetString("maintenance-requirements")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return maintenanceRequirementPartsListOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		NoAuth:                  noAuth,
		Limit:                   limit,
		Offset:                  offset,
		Sort:                    sort,
		Broker:                  broker,
		EquipmentClassification: equipmentClassification,
		Make:                    make,
		Model:                   model,
		Year:                    year,
		IsTemplate:              isTemplate,
		MaintenanceRequirements: maintenanceRequirements,
	}, nil
}

func buildMaintenanceRequirementPartRows(resp jsonAPIResponse) []maintenanceRequirementPartRow {
	rows := make([]maintenanceRequirementPartRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := maintenanceRequirementPartRow{
			ID:         resource.ID,
			PartNumber: stringAttr(resource.Attributes, "part-number"),
			Name:       stringAttr(resource.Attributes, "name"),
			Make:       stringAttr(resource.Attributes, "make"),
			Model:      stringAttr(resource.Attributes, "model"),
			Year:       stringAttr(resource.Attributes, "year"),
			IsTemplate: boolAttr(resource.Attributes, "is-template"),
		}

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["equipment-classification"]; ok && rel.Data != nil {
			row.EquipmentClassificationID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderMaintenanceRequirementPartsTable(cmd *cobra.Command, rows []maintenanceRequirementPartRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No maintenance requirement parts found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tPART NUMBER\tMAKE\tMODEL\tYEAR\tTEMPLATE\tBROKER\tEQUIP CLASS")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 30),
			truncateString(row.PartNumber, 20),
			row.Make,
			row.Model,
			row.Year,
			formatYesNo(row.IsTemplate),
			row.BrokerID,
			row.EquipmentClassificationID,
		)
	}
	return writer.Flush()
}
