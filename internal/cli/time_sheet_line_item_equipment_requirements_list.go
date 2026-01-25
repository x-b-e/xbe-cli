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

type timeSheetLineItemEquipmentRequirementsListOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	NoAuth               bool
	Limit                int
	Offset               int
	Sort                 string
	TimeSheetLineItem    string
	EquipmentRequirement string
}

func newTimeSheetLineItemEquipmentRequirementsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List time sheet line item equipment requirements",
		Long: `List time sheet line item equipment requirements with filtering and pagination.

Output Columns:
  ID                     Requirement link identifier
  TIME SHEET LINE ITEM   Time sheet line item ID
  EQUIPMENT REQUIREMENT  Equipment requirement ID
  PRIMARY                Primary requirement indicator

Filters:
  --time-sheet-line-item     Filter by time sheet line item ID
  --equipment-requirement    Filter by equipment requirement ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List time sheet line item equipment requirements
  xbe view time-sheet-line-item-equipment-requirements list

  # Filter by time sheet line item
  xbe view time-sheet-line-item-equipment-requirements list --time-sheet-line-item 123

  # Filter by equipment requirement
  xbe view time-sheet-line-item-equipment-requirements list --equipment-requirement 456

  # JSON output
  xbe view time-sheet-line-item-equipment-requirements list --json`,
		Args: cobra.NoArgs,
		RunE: runTimeSheetLineItemEquipmentRequirementsList,
	}
	initTimeSheetLineItemEquipmentRequirementsListFlags(cmd)
	return cmd
}

func init() {
	timeSheetLineItemEquipmentRequirementsCmd.AddCommand(newTimeSheetLineItemEquipmentRequirementsListCmd())
}

func initTimeSheetLineItemEquipmentRequirementsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("time-sheet-line-item", "", "Filter by time sheet line item ID")
	cmd.Flags().String("equipment-requirement", "", "Filter by equipment requirement ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeSheetLineItemEquipmentRequirementsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTimeSheetLineItemEquipmentRequirementsListOptions(cmd)
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
	query.Set("fields[time-sheet-line-item-equipment-requirements]", "is-primary,time-sheet-line-item,equipment-requirement")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[time-sheet-line-item]", opts.TimeSheetLineItem)
	setFilterIfPresent(query, "filter[equipment-requirement]", opts.EquipmentRequirement)

	body, _, err := client.Get(cmd.Context(), "/v1/time-sheet-line-item-equipment-requirements", query)
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

	rows := buildTimeSheetLineItemEquipmentRequirementRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTimeSheetLineItemEquipmentRequirementsTable(cmd, rows)
}

func parseTimeSheetLineItemEquipmentRequirementsListOptions(cmd *cobra.Command) (timeSheetLineItemEquipmentRequirementsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	timeSheetLineItem, _ := cmd.Flags().GetString("time-sheet-line-item")
	equipmentRequirement, _ := cmd.Flags().GetString("equipment-requirement")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeSheetLineItemEquipmentRequirementsListOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		NoAuth:               noAuth,
		Limit:                limit,
		Offset:               offset,
		Sort:                 sort,
		TimeSheetLineItem:    timeSheetLineItem,
		EquipmentRequirement: equipmentRequirement,
	}, nil
}

type timeSheetLineItemEquipmentRequirementRow struct {
	ID                     string `json:"id"`
	TimeSheetLineItemID    string `json:"time_sheet_line_item_id,omitempty"`
	EquipmentRequirementID string `json:"equipment_requirement_id,omitempty"`
	IsPrimary              bool   `json:"is_primary"`
}

func buildTimeSheetLineItemEquipmentRequirementRows(resp jsonAPIResponse) []timeSheetLineItemEquipmentRequirementRow {
	rows := make([]timeSheetLineItemEquipmentRequirementRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := timeSheetLineItemEquipmentRequirementRow{
			ID:        resource.ID,
			IsPrimary: boolAttr(resource.Attributes, "is-primary"),
		}

		if rel, ok := resource.Relationships["time-sheet-line-item"]; ok && rel.Data != nil {
			row.TimeSheetLineItemID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["equipment-requirement"]; ok && rel.Data != nil {
			row.EquipmentRequirementID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderTimeSheetLineItemEquipmentRequirementsTable(cmd *cobra.Command, rows []timeSheetLineItemEquipmentRequirementRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No time sheet line item equipment requirements found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTIME SHEET LINE ITEM\tEQUIPMENT REQUIREMENT\tPRIMARY")
	for _, row := range rows {
		primary := "no"
		if row.IsPrimary {
			primary = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			row.TimeSheetLineItemID,
			row.EquipmentRequirementID,
			primary,
		)
	}
	return writer.Flush()
}
