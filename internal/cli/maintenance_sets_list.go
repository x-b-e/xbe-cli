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

type maintenanceSetsListOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	NoAuth         bool
	Limit          int
	Offset         int
	Me             bool
	BusinessUnitID string
	EquipmentID    string
	Status         string
	Type           string
	Sort           string
}

func newMaintenanceSetsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List maintenance requirement sets",
		Long: `List maintenance requirement sets with filtering and pagination.

Returns a list of requirement sets that group maintenance requirements
for equipment.

Output Columns (table format):
  ID            Unique set identifier
  STATUS        Current status (editing, ready_for_work, on_hold, in_progress, completed)
  TYPE          Set type (inspection, maintenance)
  NAME          Set name
  EQUIPMENT     Associated equipment
  COMPLETION    Completion progress (completed/total)

Filtering:
  --me               Show sets for my business units (complex ownership check)
  --bu-id            Filter by business unit ID (complex ownership check)
  --equipment-id     Filter by equipment ID
  --status           Filter by status (comma-separated)
  --type             Filter by type (inspection, maintenance)

Note: The --me and --bu-id flags use client-side filtering to properly match
sets that belong to a BU via equipment classification and requirements.

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.
  Default: -created-at (newest first)`,
		Example: `  # List all sets
  xbe view maintenance sets list

  # List sets for my business units
  xbe view maintenance sets list --me

  # Filter by equipment
  xbe view maintenance sets list --equipment-id 123

  # Filter by business unit (complex ownership check)
  xbe view maintenance sets list --bu-id 456

  # Filter by status
  xbe view maintenance sets list --status in_progress

  # Filter by multiple statuses
  xbe view maintenance sets list --status editing,ready_for_work

  # Filter by type
  xbe view maintenance sets list --type inspection

  # Combine filters
  xbe view maintenance sets list --me --status in_progress

  # Paginate results
  xbe view maintenance sets list --limit 50 --offset 100

  # Output as JSON
  xbe view maintenance sets list --json`,
		RunE: runMaintenanceSetsList,
	}
	initMaintenanceSetsListFlags(cmd)
	return cmd
}

func init() {
	maintenanceSetsCmd.AddCommand(newMaintenanceSetsListCmd())
}

func initMaintenanceSetsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Bool("me", false, "Show sets for my business units")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("bu-id", "", "Filter by business unit ID")
	cmd.Flags().String("equipment-id", "", "Filter by equipment ID")
	cmd.Flags().String("status", "", "Filter by status (comma-separated: editing,ready_for_work,on_hold,in_progress,completed)")
	cmd.Flags().String("type", "", "Filter by type (inspection, maintenance)")
	cmd.Flags().String("sort", "", "Sort order (default: -created-at)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaintenanceSetsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaintenanceSetsListOptions(cmd)
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
	// Include requirements with equipment for client-side filtering
	query.Set("include", "maintenance-requirements,maintenance-requirements.equipment,equipments,equipment-classification")

	// When doing client-side filtering, we need to fetch more records
	if buContext != nil {
		query.Set("page[limit]", "500")
	} else {
		if opts.Limit > 0 {
			query.Set("page[limit]", strconv.Itoa(opts.Limit))
		}
		if opts.Offset > 0 {
			query.Set("page[offset]", strconv.Itoa(opts.Offset))
		}
	}

	// Apply direct API filters
	setFilterIfPresent(query, "filter[equipment]", opts.EquipmentID)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[maintenance_type]", opts.Type)

	// Apply sort
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	} else {
		query.Set("sort", "-created-at")
	}

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

	// Apply client-side BU filtering if needed
	if buContext != nil {
		resp = filterSetsByBUContext(resp, buContext, opts.Limit, opts.Offset)
	}

	if opts.JSON {
		rows := buildSetRows(resp)
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderSetsList(cmd, resp)
}

func parseMaintenanceSetsListOptions(cmd *cobra.Command) (maintenanceSetsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	me, _ := cmd.Flags().GetBool("me")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	businessUnitID, _ := cmd.Flags().GetString("bu-id")
	equipmentID, _ := cmd.Flags().GetString("equipment-id")
	status, _ := cmd.Flags().GetString("status")
	typeFilter, _ := cmd.Flags().GetString("type")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return maintenanceSetsListOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		NoAuth:         noAuth,
		Limit:          limit,
		Offset:         offset,
		Me:             me,
		BusinessUnitID: businessUnitID,
		EquipmentID:    equipmentID,
		Status:         status,
		Type:           typeFilter,
		Sort:           sort,
	}, nil
}

// filterSetsByBUContext applies client-side filtering using the BU equipment context
func filterSetsByBUContext(resp jsonAPIResponse, ctx *BUEquipmentContext, limit, offset int) jsonAPIResponse {
	// Build included map for lookups
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	filtered := make([]jsonAPIResource, 0)
	for _, set := range resp.Data {
		if canAccessRequirementSet(set, included, ctx) {
			filtered = append(filtered, set)
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

func buildSetRows(resp jsonAPIResponse) []setRow {
	// Build included map for lookups
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]setRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes

		row := setRow{
			ID:     resource.ID,
			Status: stringAttr(attrs, "status"),
			Type:   stringAttr(attrs, "maintenance-type"),
			Name:   strings.TrimSpace(stringAttr(attrs, "template-name")),
		}

		// Get equipment info (may be array)
		if rel, ok := resource.Relationships["equipments"]; ok && rel.raw != nil {
			var refs []jsonAPIResourceIdentifier
			if err := json.Unmarshal(rel.raw, &refs); err == nil && len(refs) > 0 {
				row.EquipmentID = refs[0].ID
				key := resourceKey(refs[0].Type, refs[0].ID)
				if inc, ok := included[key]; ok {
					row.Equipment = firstNonEmpty(
						stringAttr(inc.Attributes, "name"),
						stringAttr(inc.Attributes, "equipment-number"),
						refs[0].ID,
					)
				}
			}
		} else if rel, ok := resource.Relationships["equipment"]; ok && rel.Data != nil {
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

		// Count requirements for completion stats
		if rel, ok := resource.Relationships["maintenance-requirements"]; ok && rel.raw != nil {
			var refs []jsonAPIResourceIdentifier
			if err := json.Unmarshal(rel.raw, &refs); err == nil {
				row.TotalCount = len(refs)
				for _, ref := range refs {
					key := resourceKey(ref.Type, ref.ID)
					if inc, ok := included[key]; ok {
						if stringAttr(inc.Attributes, "status") == "completed" {
							row.CompletionCount++
						}
					}
				}
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderSetsList(cmd *cobra.Command, resp jsonAPIResponse) error {
	rows := buildSetRows(resp)
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No requirement sets found.")
		return nil
	}

	const nameMax = 30
	const equipmentMax = 20

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tTYPE\tNAME\tEQUIPMENT\tCOMPLETION")
	for _, row := range rows {
		completion := fmt.Sprintf("%d/%d", row.CompletionCount, row.TotalCount)
		if row.TotalCount == 0 {
			completion = "-"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			row.Type,
			truncateString(row.Name, nameMax),
			truncateString(row.Equipment, equipmentMax),
			completion,
		)
	}
	return writer.Flush()
}
