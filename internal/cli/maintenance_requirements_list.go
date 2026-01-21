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

type maintenanceRequirementsListOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	NoAuth         bool
	Limit          int
	Offset         int
	Me             bool
	BusinessUnitID string
	SetID          string
	EquipmentID    string
	Status         string
	Templates      bool
	Sort           string
}

func newMaintenanceRequirementsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List maintenance requirements",
		Long: `List maintenance requirements with filtering and pagination.

Returns a list of maintenance requirements that define specific maintenance
tasks to be performed on equipment.

Output Columns (table format):
  ID            Unique requirement identifier
  STATUS        Current status (pending, on_hold, in_progress, completed)
  DESCRIPTION   Requirement description (truncated)
  EQUIPMENT     Associated equipment
  SET           Parent requirement set

Filtering:
  --me              Show requirements for my business units' equipment
  --bu-id           Filter by business unit ID (via equipment)
  --set-id          Filter by requirement set ID
  --equipment-id    Filter by equipment ID
  --status          Filter by status (comma-separated: pending,on_hold,in_progress,completed)
  --templates       Show only template requirements

Note: The --me and --bu-id flags filter requirements by equipment ownership.

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.
  Default: -created-at (newest first)`,
		Example: `  # List all requirements
  xbe view maintenance requirements list

  # List requirements for my business units
  xbe view maintenance requirements list --me

  # Filter by set
  xbe view maintenance requirements list --set-id 123

  # Filter by equipment
  xbe view maintenance requirements list --equipment-id 456

  # Filter by business unit (via equipment ownership)
  xbe view maintenance requirements list --bu-id 789

  # Filter by status
  xbe view maintenance requirements list --status pending

  # Filter by multiple statuses
  xbe view maintenance requirements list --status pending,in_progress

  # Show only templates
  xbe view maintenance requirements list --templates

  # Combine filters
  xbe view maintenance requirements list --me --status in_progress

  # Paginate results
  xbe view maintenance requirements list --limit 50 --offset 100

  # Output as JSON
  xbe view maintenance requirements list --json`,
		RunE: runMaintenanceRequirementsList,
	}
	initMaintenanceRequirementsListFlags(cmd)
	return cmd
}

func init() {
	maintenanceRequirementsCmd.AddCommand(newMaintenanceRequirementsListCmd())
}

func initMaintenanceRequirementsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Bool("me", false, "Show requirements for my business units")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("bu-id", "", "Filter by business unit ID (via equipment)")
	cmd.Flags().String("set-id", "", "Filter by requirement set ID")
	cmd.Flags().String("equipment-id", "", "Filter by equipment ID")
	cmd.Flags().String("status", "", "Filter by status (comma-separated: pending,on_hold,in_progress,completed)")
	cmd.Flags().Bool("templates", false, "Show only template requirements")
	cmd.Flags().String("sort", "", "Sort order (default: -created-at)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaintenanceRequirementsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaintenanceRequirementsListOptions(cmd)
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

	// For requirements, we filter via equipment IDs
	// Simple ownership: requirement → equipment → BU
	var equipmentFilter string
	if opts.Me {
		if opts.BusinessUnitID != "" {
			return fmt.Errorf("cannot use both --me and --bu-id")
		}
		if opts.EquipmentID != "" {
			return fmt.Errorf("cannot use both --me and --equipment-id")
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
		if len(ctx.EquipmentIDs) > 0 {
			equipmentFilter = strings.Join(ctx.EquipmentIDs, ",")
		}
	} else if opts.BusinessUnitID != "" {
		if opts.EquipmentID != "" {
			return fmt.Errorf("cannot use both --bu-id and --equipment-id")
		}
		buIDs := []string{opts.BusinessUnitID}
		ctx, err := getBUEquipmentContext(cmd, client, buIDs)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		if len(ctx.EquipmentIDs) > 0 {
			equipmentFilter = strings.Join(ctx.EquipmentIDs, ",")
		}
	} else if opts.EquipmentID != "" {
		equipmentFilter = opts.EquipmentID
	}

	query := url.Values{}
	query.Set("include", "equipment,maintenance-requirement-sets")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	// Apply filters
	setFilterIfPresent(query, "filter[maintenance_requirement_sets]", opts.SetID)
	if equipmentFilter != "" {
		query.Set("filter[equipment]", equipmentFilter)
	}
	setFilterIfPresent(query, "filter[status]", opts.Status)
	if opts.Templates {
		query.Set("filter[is_template]", "true")
	}

	// Apply sort
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	} else {
		query.Set("sort", "-created-at")
	}

	body, _, err := client.Get(cmd.Context(), "/v1/maintenance-requirements", query)
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

	if opts.JSON {
		rows := buildRequirementRows(resp)
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderRequirementsList(cmd, resp)
}

func parseMaintenanceRequirementsListOptions(cmd *cobra.Command) (maintenanceRequirementsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	me, _ := cmd.Flags().GetBool("me")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	businessUnitID, _ := cmd.Flags().GetString("bu-id")
	setID, _ := cmd.Flags().GetString("set-id")
	equipmentID, _ := cmd.Flags().GetString("equipment-id")
	status, _ := cmd.Flags().GetString("status")
	templates, _ := cmd.Flags().GetBool("templates")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return maintenanceRequirementsListOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		NoAuth:         noAuth,
		Limit:          limit,
		Offset:         offset,
		Me:             me,
		BusinessUnitID: businessUnitID,
		SetID:          setID,
		EquipmentID:    equipmentID,
		Status:         status,
		Templates:      templates,
		Sort:           sort,
	}, nil
}

func buildRequirementRows(resp jsonAPIResponse) []requirementRow {
	// Build included map for lookups
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]requirementRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes

		row := requirementRow{
			ID:          resource.ID,
			Status:      stringAttr(attrs, "status"),
			Description: strings.TrimSpace(stringAttr(attrs, "description")),
			DueDate:     formatDate(stringAttr(attrs, "due-on")),
			IsTemplate:  boolAttr(attrs, "is-template"),
		}

		// Get equipment info
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

		// Get set info (may be an array relationship)
		if rel, ok := resource.Relationships["maintenance-requirement-sets"]; ok && rel.raw != nil {
			var refs []jsonAPIResourceIdentifier
			if err := json.Unmarshal(rel.raw, &refs); err == nil && len(refs) > 0 {
				row.SetID = refs[0].ID
				key := resourceKey(refs[0].Type, refs[0].ID)
				if inc, ok := included[key]; ok {
					row.SetName = stringAttr(inc.Attributes, "template-name")
				}
			}
		} else if rel, ok := resource.Relationships["maintenance-requirement-sets"]; ok && rel.Data != nil {
			row.SetID = rel.Data.ID
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if inc, ok := included[key]; ok {
				row.SetName = stringAttr(inc.Attributes, "template-name")
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderRequirementsList(cmd *cobra.Command, resp jsonAPIResponse) error {
	rows := buildRequirementRows(resp)
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No requirements found.")
		return nil
	}

	const descMax = 35
	const equipmentMax = 20
	const setMax = 20

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tDESCRIPTION\tEQUIPMENT\tSET")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			truncateString(row.Description, descMax),
			truncateString(row.Equipment, equipmentMax),
			truncateString(row.SetName, setMax),
		)
	}
	return writer.Flush()
}
