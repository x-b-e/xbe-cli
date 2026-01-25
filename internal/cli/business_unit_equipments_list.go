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

type businessUnitEquipmentsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	BusinessUnit string
	Equipment    string
}

type businessUnitEquipmentRow struct {
	ID                     string `json:"id"`
	BusinessUnitID         string `json:"business_unit_id,omitempty"`
	BusinessUnitName       string `json:"business_unit_name,omitempty"`
	BusinessUnitExternalID string `json:"business_unit_external_id,omitempty"`
	EquipmentID            string `json:"equipment_id,omitempty"`
	EquipmentNickname      string `json:"equipment_nickname,omitempty"`
	EquipmentSerialNumber  string `json:"equipment_serial_number,omitempty"`
}

func newBusinessUnitEquipmentsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List business unit equipment links",
		Long: `List business unit equipment links with filtering and pagination.

Output Columns:
  ID             Business unit equipment link identifier
  BUSINESS UNIT  Business unit name or ID
  EQUIPMENT      Equipment nickname or ID
  SERIAL         Equipment serial number

Filters:
  --business-unit  Filter by business unit ID
  --equipment      Filter by equipment ID

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List business unit equipment links
  xbe view business-unit-equipments list

  # Filter by business unit
  xbe view business-unit-equipments list --business-unit 123

  # Filter by equipment
  xbe view business-unit-equipments list --equipment 456

  # Output as JSON
  xbe view business-unit-equipments list --json`,
		Args: cobra.NoArgs,
		RunE: runBusinessUnitEquipmentsList,
	}
	initBusinessUnitEquipmentsListFlags(cmd)
	return cmd
}

func init() {
	businessUnitEquipmentsCmd.AddCommand(newBusinessUnitEquipmentsListCmd())
}

func initBusinessUnitEquipmentsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("business-unit", "", "Filter by business unit ID")
	cmd.Flags().String("equipment", "", "Filter by equipment ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBusinessUnitEquipmentsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseBusinessUnitEquipmentsListOptions(cmd)
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
	query.Set("fields[business-unit-equipments]", "business-unit,equipment")
	query.Set("include", "business-unit,equipment")
	query.Set("fields[business-units]", "company-name,external-id")
	query.Set("fields[equipment]", "nickname,serial-number")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[business-unit]", opts.BusinessUnit)
	setFilterIfPresent(query, "filter[equipment]", opts.Equipment)

	body, _, err := client.Get(cmd.Context(), "/v1/business-unit-equipments", query)
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

	rows := buildBusinessUnitEquipmentRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderBusinessUnitEquipmentsTable(cmd, rows)
}

func parseBusinessUnitEquipmentsListOptions(cmd *cobra.Command) (businessUnitEquipmentsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	businessUnit, _ := cmd.Flags().GetString("business-unit")
	equipment, _ := cmd.Flags().GetString("equipment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return businessUnitEquipmentsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		BusinessUnit: businessUnit,
		Equipment:    equipment,
	}, nil
}

func buildBusinessUnitEquipmentRows(resp jsonAPIResponse) []businessUnitEquipmentRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]businessUnitEquipmentRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildBusinessUnitEquipmentRow(resource, included))
	}
	return rows
}

func businessUnitEquipmentRowFromSingle(resp jsonAPISingleResponse) businessUnitEquipmentRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildBusinessUnitEquipmentRow(resp.Data, included)
}

func buildBusinessUnitEquipmentRow(resource jsonAPIResource, included map[string]jsonAPIResource) businessUnitEquipmentRow {
	row := businessUnitEquipmentRow{
		ID: resource.ID,
	}

	if rel, ok := resource.Relationships["business-unit"]; ok && rel.Data != nil {
		row.BusinessUnitID = rel.Data.ID
		if businessUnit, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.BusinessUnitName = stringAttr(businessUnit.Attributes, "company-name")
			row.BusinessUnitExternalID = stringAttr(businessUnit.Attributes, "external-id")
		}
	}

	if rel, ok := resource.Relationships["equipment"]; ok && rel.Data != nil {
		row.EquipmentID = rel.Data.ID
		if equipment, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.EquipmentNickname = stringAttr(equipment.Attributes, "nickname")
			row.EquipmentSerialNumber = stringAttr(equipment.Attributes, "serial-number")
		}
	}

	return row
}

func renderBusinessUnitEquipmentsTable(cmd *cobra.Command, rows []businessUnitEquipmentRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No business unit equipments found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tBUSINESS UNIT\tEQUIPMENT\tSERIAL")
	for _, row := range rows {
		businessUnit := firstNonEmpty(row.BusinessUnitName, row.BusinessUnitID)
		equipment := firstNonEmpty(row.EquipmentNickname, row.EquipmentID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(businessUnit, 24),
			truncateString(equipment, 24),
			truncateString(row.EquipmentSerialNumber, 18),
		)
	}
	return writer.Flush()
}
