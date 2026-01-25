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

type equipmentListOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	NoAuth                  bool
	Limit                   int
	Offset                  int
	EquipmentClassification string
	OrganizationType        string
	OrganizationID          string
	IsActive                string
	MobilizationMethod      string
	IsRented                string
	IsAvailable             string
	Broker                  string
	Tractor                 string
	Trailer                 string
	NicknameLike            string
	BusinessUnit            string
	Search                  string
}

func newEquipmentListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List equipment",
		Long: `List equipment with filtering and pagination.

Equipment represents tracked assets like tools, machines, and other items.

Output Columns:
  ID              Equipment identifier
  NICKNAME        Equipment nickname
  SERIAL          Serial number
  CLASSIFICATION  Equipment classification ID
  ORG TYPE        Organization type
  ORG ID          Organization ID
  ACTIVE          Whether equipment is active

Filters:
  --equipment-classification   Filter by equipment classification ID
  --organization-type          Filter by organization type (e.g., Broker, Customer)
  --organization-id            Filter by organization ID
  --is-active                  Filter by active status (true/false)
  --mobilization-method        Filter by mobilization method
  --is-rented                  Filter by rented status (true/false)
  --is-available               Filter by availability status (true/false)
  --broker                     Filter by broker ID
  --tractor                    Filter by tractor ID
  --trailer                    Filter by trailer ID
  --nickname-like              Filter by nickname (partial match)
  --business-unit              Filter by business unit ID
  --search                     Search equipment`,
		Example: `  # List all equipment
  xbe view equipment list

  # Filter by classification
  xbe view equipment list --equipment-classification 123

  # Filter by organization
  xbe view equipment list --organization-type Broker --organization-id 456

  # Filter by active status
  xbe view equipment list --is-active true

  # Search by nickname
  xbe view equipment list --nickname-like "Excavator"

  # Output as JSON
  xbe view equipment list --json`,
		RunE: runEquipmentList,
	}
	initEquipmentListFlags(cmd)
	return cmd
}

func init() {
	equipmentCmd.AddCommand(newEquipmentListCmd())
}

func initEquipmentListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("equipment-classification", "", "Filter by equipment classification ID")
	cmd.Flags().String("organization-type", "", "Filter by organization type (e.g., Broker, Customer)")
	cmd.Flags().String("organization-id", "", "Filter by organization ID")
	cmd.Flags().String("is-active", "", "Filter by active status (true/false)")
	cmd.Flags().String("mobilization-method", "", "Filter by mobilization method")
	cmd.Flags().String("is-rented", "", "Filter by rented status (true/false)")
	cmd.Flags().String("is-available", "", "Filter by availability status (true/false)")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("tractor", "", "Filter by tractor ID")
	cmd.Flags().String("trailer", "", "Filter by trailer ID")
	cmd.Flags().String("nickname-like", "", "Filter by nickname (partial match)")
	cmd.Flags().String("business-unit", "", "Filter by business unit ID")
	cmd.Flags().String("search", "", "Search equipment")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentList(cmd *cobra.Command, _ []string) error {
	opts, err := parseEquipmentListOptions(cmd)
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
	query.Set("fields[equipment]", "nickname,serial-number,is-active,equipment-classification,organization")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[equipment_classification]", opts.EquipmentClassification)
	if opts.OrganizationType != "" && opts.OrganizationID != "" {
		query.Set("filter[by_organization]", opts.OrganizationType+"|"+opts.OrganizationID)
	}
	setFilterIfPresent(query, "filter[is_active]", opts.IsActive)
	setFilterIfPresent(query, "filter[mobilization_method]", opts.MobilizationMethod)
	setFilterIfPresent(query, "filter[is_rented]", opts.IsRented)
	setFilterIfPresent(query, "filter[is_available]", opts.IsAvailable)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[tractor]", opts.Tractor)
	setFilterIfPresent(query, "filter[trailer]", opts.Trailer)
	setFilterIfPresent(query, "filter[nickname_like]", opts.NicknameLike)
	setFilterIfPresent(query, "filter[business_unit]", opts.BusinessUnit)
	setFilterIfPresent(query, "filter[q]", opts.Search)

	body, _, err := client.Get(cmd.Context(), "/v1/equipment", query)
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

	rows := buildEquipmentRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderEquipmentTable(cmd, rows)
}

func parseEquipmentListOptions(cmd *cobra.Command) (equipmentListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	equipmentClassification, _ := cmd.Flags().GetString("equipment-classification")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	isActive, _ := cmd.Flags().GetString("is-active")
	mobilizationMethod, _ := cmd.Flags().GetString("mobilization-method")
	isRented, _ := cmd.Flags().GetString("is-rented")
	isAvailable, _ := cmd.Flags().GetString("is-available")
	broker, _ := cmd.Flags().GetString("broker")
	tractor, _ := cmd.Flags().GetString("tractor")
	trailer, _ := cmd.Flags().GetString("trailer")
	nicknameLike, _ := cmd.Flags().GetString("nickname-like")
	businessUnit, _ := cmd.Flags().GetString("business-unit")
	search, _ := cmd.Flags().GetString("search")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentListOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		NoAuth:                  noAuth,
		Limit:                   limit,
		Offset:                  offset,
		EquipmentClassification: equipmentClassification,
		OrganizationType:        organizationType,
		OrganizationID:          organizationID,
		IsActive:                isActive,
		MobilizationMethod:      mobilizationMethod,
		IsRented:                isRented,
		IsAvailable:             isAvailable,
		Broker:                  broker,
		Tractor:                 tractor,
		Trailer:                 trailer,
		NicknameLike:            nicknameLike,
		BusinessUnit:            businessUnit,
		Search:                  search,
	}, nil
}

type equipmentRow struct {
	ID                        string `json:"id"`
	Nickname                  string `json:"nickname,omitempty"`
	SerialNumber              string `json:"serial_number,omitempty"`
	IsActive                  bool   `json:"is_active,omitempty"`
	EquipmentClassificationID string `json:"equipment_classification_id,omitempty"`
	OrganizationType          string `json:"organization_type,omitempty"`
	OrganizationID            string `json:"organization_id,omitempty"`
}

func buildEquipmentRows(resp jsonAPIResponse) []equipmentRow {
	rows := make([]equipmentRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := equipmentRow{
			ID:           resource.ID,
			Nickname:     stringAttr(resource.Attributes, "nickname"),
			SerialNumber: stringAttr(resource.Attributes, "serial-number"),
			IsActive:     boolAttr(resource.Attributes, "is-active"),
		}

		if rel, ok := resource.Relationships["equipment-classification"]; ok && rel.Data != nil {
			row.EquipmentClassificationID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
			row.OrganizationType = rel.Data.Type
			row.OrganizationID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderEquipmentTable(cmd *cobra.Command, rows []equipmentRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No equipment found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNICKNAME\tSERIAL\tCLASSIFICATION\tORG TYPE\tORG ID\tACTIVE")
	for _, row := range rows {
		active := "no"
		if row.IsActive {
			active = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Nickname, 25),
			truncateString(row.SerialNumber, 15),
			row.EquipmentClassificationID,
			row.OrganizationType,
			row.OrganizationID,
			active,
		)
	}
	return writer.Flush()
}
