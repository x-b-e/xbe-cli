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

type maintenanceRequirementMaintenanceRequirementPartsListOptions struct {
	BaseURL                    string
	Token                      string
	JSON                       bool
	NoAuth                     bool
	Limit                      int
	Offset                     int
	Sort                       string
	MaintenanceRequirement     string
	MaintenanceRequirementPart string
}

type maintenanceRequirementMaintenanceRequirementPartRow struct {
	ID                           string `json:"id"`
	MaintenanceRequirementID     string `json:"maintenance_requirement_id,omitempty"`
	MaintenanceRequirement       string `json:"maintenance_requirement,omitempty"`
	MaintenanceRequirementPartID string `json:"maintenance_requirement_part_id,omitempty"`
	PartName                     string `json:"part_name,omitempty"`
	PartNumber                   string `json:"part_part_number,omitempty"`
	Quantity                     string `json:"quantity,omitempty"`
	UnitCost                     string `json:"unit_cost,omitempty"`
	TotalCost                    string `json:"total_cost,omitempty"`
	Source                       string `json:"source,omitempty"`
}

func newMaintenanceRequirementMaintenanceRequirementPartsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List maintenance requirement parts",
		Long: `List maintenance requirement parts with filtering and pagination.

Output Columns:
  ID           Link identifier
  REQUIREMENT  Maintenance requirement (template name or description)
  PART         Part name or part number
  QTY          Required quantity
  UNIT COST    Unit cost
  TOTAL COST   Total cost
  SOURCE       Part source

Filters:
  --maintenance-requirement       Filter by maintenance requirement ID
  --maintenance-requirement-part  Filter by maintenance requirement part ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List maintenance requirement parts
  xbe view maintenance-requirement-maintenance-requirement-parts list

  # Filter by maintenance requirement
  xbe view maintenance-requirement-maintenance-requirement-parts list --maintenance-requirement 123

  # Filter by maintenance requirement part
  xbe view maintenance-requirement-maintenance-requirement-parts list --maintenance-requirement-part 456

  # JSON output
  xbe view maintenance-requirement-maintenance-requirement-parts list --json`,
		Args: cobra.NoArgs,
		RunE: runMaintenanceRequirementMaintenanceRequirementPartsList,
	}
	initMaintenanceRequirementMaintenanceRequirementPartsListFlags(cmd)
	return cmd
}

func init() {
	maintenanceRequirementMaintenanceRequirementPartsCmd.AddCommand(newMaintenanceRequirementMaintenanceRequirementPartsListCmd())
}

func initMaintenanceRequirementMaintenanceRequirementPartsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("maintenance-requirement", "", "Filter by maintenance requirement ID")
	cmd.Flags().String("maintenance-requirement-part", "", "Filter by maintenance requirement part ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaintenanceRequirementMaintenanceRequirementPartsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaintenanceRequirementMaintenanceRequirementPartsListOptions(cmd)
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
	query.Set("fields[maintenance-requirement-maintenance-requirement-parts]", "quantity,unit-cost,source,total-cost,part-name,part-part-number,maintenance-requirement,maintenance-requirement-part")
	query.Set("fields[maintenance-requirements]", "template-name,description")
	query.Set("include", "maintenance-requirement")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[maintenance-requirement]", opts.MaintenanceRequirement)
	setFilterIfPresent(query, "filter[maintenance-requirement-part]", opts.MaintenanceRequirementPart)

	body, _, err := client.Get(cmd.Context(), "/v1/maintenance-requirement-maintenance-requirement-parts", query)
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

	rows := buildMaintenanceRequirementMaintenanceRequirementPartRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaintenanceRequirementMaintenanceRequirementPartsTable(cmd, rows)
}

func parseMaintenanceRequirementMaintenanceRequirementPartsListOptions(cmd *cobra.Command) (maintenanceRequirementMaintenanceRequirementPartsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	maintenanceRequirement, _ := cmd.Flags().GetString("maintenance-requirement")
	maintenanceRequirementPart, _ := cmd.Flags().GetString("maintenance-requirement-part")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return maintenanceRequirementMaintenanceRequirementPartsListOptions{
		BaseURL:                    baseURL,
		Token:                      token,
		JSON:                       jsonOut,
		NoAuth:                     noAuth,
		Limit:                      limit,
		Offset:                     offset,
		Sort:                       sort,
		MaintenanceRequirement:     maintenanceRequirement,
		MaintenanceRequirementPart: maintenanceRequirementPart,
	}, nil
}

func buildMaintenanceRequirementMaintenanceRequirementPartRows(resp jsonAPIResponse) []maintenanceRequirementMaintenanceRequirementPartRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]maintenanceRequirementMaintenanceRequirementPartRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := maintenanceRequirementMaintenanceRequirementPartRow{
			ID:         resource.ID,
			Quantity:   stringAttr(resource.Attributes, "quantity"),
			UnitCost:   stringAttr(resource.Attributes, "unit-cost"),
			TotalCost:  stringAttr(resource.Attributes, "total-cost"),
			Source:     stringAttr(resource.Attributes, "source"),
			PartName:   stringAttr(resource.Attributes, "part-name"),
			PartNumber: stringAttr(resource.Attributes, "part-part-number"),
		}

		if rel, ok := resource.Relationships["maintenance-requirement"]; ok && rel.Data != nil {
			row.MaintenanceRequirementID = rel.Data.ID
			if req, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				templateName := stringAttr(req.Attributes, "template-name")
				description := stringAttr(req.Attributes, "description")
				row.MaintenanceRequirement = firstNonEmpty(templateName, description)
			}
		}
		if rel, ok := resource.Relationships["maintenance-requirement-part"]; ok && rel.Data != nil {
			row.MaintenanceRequirementPartID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func buildMaintenanceRequirementMaintenanceRequirementPartRowFromSingle(resp jsonAPISingleResponse) maintenanceRequirementMaintenanceRequirementPartRow {
	rows := buildMaintenanceRequirementMaintenanceRequirementPartRows(jsonAPIResponse{
		Data:     []jsonAPIResource{resp.Data},
		Included: resp.Included,
	})
	if len(rows) == 0 {
		return maintenanceRequirementMaintenanceRequirementPartRow{}
	}
	return rows[0]
}

func renderMaintenanceRequirementMaintenanceRequirementPartsTable(cmd *cobra.Command, rows []maintenanceRequirementMaintenanceRequirementPartRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No maintenance requirement parts found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tREQUIREMENT\tPART\tQTY\tUNIT COST\tTOTAL COST\tSOURCE")
	for _, row := range rows {
		requirement := firstNonEmpty(row.MaintenanceRequirement, row.MaintenanceRequirementID)
		partLabel := formatMaintenanceRequirementPartLabel(row.PartName, row.PartNumber)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(requirement, 30),
			truncateString(partLabel, 30),
			row.Quantity,
			row.UnitCost,
			row.TotalCost,
			row.Source,
		)
	}
	return writer.Flush()
}

func formatMaintenanceRequirementPartLabel(name, number string) string {
	name = strings.TrimSpace(name)
	number = strings.TrimSpace(number)
	if name != "" && number != "" {
		return fmt.Sprintf("%s (%s)", name, number)
	}
	return firstNonEmpty(name, number)
}
