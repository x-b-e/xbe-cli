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

type workOrdersListOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	NoAuth         bool
	Limit          int
	Offset         int
	Me             bool
	BusinessUnitID string
	Status         string
	Priority       string
	DueBefore      string
	DueAfter       string
	Sort           string
}

type workOrderRow struct {
	ID             string `json:"id"`
	Status         string `json:"status,omitempty"`
	Priority       string `json:"priority,omitempty"`
	DueDate        string `json:"due_date,omitempty"`
	BusinessUnitID string `json:"business_unit_id,omitempty"`
	BusinessUnit   string `json:"business_unit,omitempty"`
	ServiceSiteID  string `json:"service_site_id,omitempty"`
	ServiceSite    string `json:"service_site,omitempty"`
	SetCount       int    `json:"set_count,omitempty"`
}

func newWorkOrdersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List work orders",
		Long: `List work orders with filtering and pagination.

Returns a list of work orders that group maintenance requirement sets.

Output Columns (table format):
  ID              Unique work order identifier
  STATUS          Current status (editing, ready_for_work, in_progress, on_hold, completed)
  PRIORITY        Priority level (urgent, high, normal, low)
  DUE_DATE        Due date for completion
  BUSINESS_UNIT   Responsible business unit
  SERVICE_SITE    Service site location
  SETS            Number of requirement sets

Filtering:
  --me                  Show work orders for my business units
  --bu-id               Filter by business unit ID (responsible party)
  --status              Filter by status (comma-separated)
  --priority            Filter by priority (urgent, high, normal, low)
  --due-before          Filter by due date (before, YYYY-MM-DD)
  --due-after           Filter by due date (after, YYYY-MM-DD)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.
  Default: -created-at (newest first)`,
		Example: `  # List all work orders
  xbe view work-orders list

  # List work orders for my business units
  xbe view work-orders list --me

  # Filter by business unit
  xbe view work-orders list --bu-id 123

  # Filter by status
  xbe view work-orders list --status in_progress

  # Filter by multiple statuses
  xbe view work-orders list --status editing,ready_for_work

  # Filter by priority
  xbe view work-orders list --priority urgent

  # Filter by due date
  xbe view work-orders list --due-before 2025-12-31

  # Combine filters
  xbe view work-orders list --me --status in_progress --priority high

  # Paginate results
  xbe view work-orders list --limit 50 --offset 100

  # Output as JSON
  xbe view work-orders list --json`,
		RunE: runWorkOrdersList,
	}
	initWorkOrdersListFlags(cmd)
	return cmd
}

func init() {
	workOrdersCmd.AddCommand(newWorkOrdersListCmd())
}

func initWorkOrdersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Bool("me", false, "Show work orders for my business units")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("bu-id", "", "Filter by business unit ID (responsible party)")
	cmd.Flags().String("status", "", "Filter by status (comma-separated: editing,ready_for_work,in_progress,on_hold,completed)")
	cmd.Flags().String("priority", "", "Filter by priority (urgent, high, normal, low)")
	cmd.Flags().String("due-before", "", "Filter by due date (before, YYYY-MM-DD)")
	cmd.Flags().String("due-after", "", "Filter by due date (after, YYYY-MM-DD)")
	cmd.Flags().String("sort", "", "Sort order (default: -created-at)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runWorkOrdersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseWorkOrdersListOptions(cmd)
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

	// If --me flag is set, get the user's business unit IDs
	var buIDs []string
	if opts.Me {
		if opts.BusinessUnitID != "" {
			return fmt.Errorf("cannot use both --me and --bu-id")
		}
		ids, err := getCurrentUserBusinessUnitIDs(cmd, client)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		buIDs = ids
	} else if opts.BusinessUnitID != "" {
		buIDs = []string{opts.BusinessUnitID}
	}

	query := url.Values{}
	query.Set("include", "responsible-party,service-site,maintenance-requirement-sets")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	// Apply filters
	if len(buIDs) > 0 {
		query.Set("filter[responsible_party]", strings.Join(buIDs, ","))
	}
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[priority]", opts.Priority)
	setFilterIfPresent(query, "filter[due_date_max]", opts.DueBefore)
	setFilterIfPresent(query, "filter[due_date_min]", opts.DueAfter)

	// Apply sort
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	} else {
		query.Set("sort", "-created-at")
	}

	body, _, err := client.Get(cmd.Context(), "/v1/work-orders", query)
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
		rows := buildWorkOrderRows(resp)
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderWorkOrdersList(cmd, resp)
}

func parseWorkOrdersListOptions(cmd *cobra.Command) (workOrdersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	me, _ := cmd.Flags().GetBool("me")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	buID, _ := cmd.Flags().GetString("bu-id")
	status, _ := cmd.Flags().GetString("status")
	priority, _ := cmd.Flags().GetString("priority")
	dueBefore, _ := cmd.Flags().GetString("due-before")
	dueAfter, _ := cmd.Flags().GetString("due-after")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return workOrdersListOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		NoAuth:         noAuth,
		Limit:          limit,
		Offset:         offset,
		Me:             me,
		BusinessUnitID: buID,
		Status:         status,
		Priority:       priority,
		DueBefore:      dueBefore,
		DueAfter:       dueAfter,
		Sort:           sort,
	}, nil
}

func buildWorkOrderRows(resp jsonAPIResponse) []workOrderRow {
	// Build included map for lookups
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]workOrderRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes

		row := workOrderRow{
			ID:       resource.ID,
			Status:   stringAttr(attrs, "status"),
			Priority: stringAttr(attrs, "priority"),
			DueDate:  formatDate(stringAttr(attrs, "due-date")),
		}

		// Get responsible party (business unit)
		if rel, ok := resource.Relationships["responsible-party"]; ok && rel.Data != nil {
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

		// Get service site
		if rel, ok := resource.Relationships["service-site"]; ok && rel.Data != nil {
			row.ServiceSiteID = rel.Data.ID
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if inc, ok := included[key]; ok {
				row.ServiceSite = firstNonEmpty(
					stringAttr(inc.Attributes, "name"),
					stringAttr(inc.Attributes, "address"),
					rel.Data.ID,
				)
			}
		}

		// Count requirement sets
		if rel, ok := resource.Relationships["maintenance-requirement-sets"]; ok && rel.raw != nil {
			var refs []jsonAPIResourceIdentifier
			if err := json.Unmarshal(rel.raw, &refs); err == nil {
				row.SetCount = len(refs)
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderWorkOrdersList(cmd *cobra.Command, resp jsonAPIResponse) error {
	rows := buildWorkOrderRows(resp)
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No work orders found.")
		return nil
	}

	const buMax = 25
	const siteMax = 20

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tPRIORITY\tDUE_DATE\tBUSINESS_UNIT\tSERVICE_SITE\tSETS")
	for _, row := range rows {
		status := row.Status
		if status == "" {
			status = "-"
		}
		priority := row.Priority
		if priority == "" {
			priority = "-"
		}
		dueDate := row.DueDate
		if dueDate == "" {
			dueDate = "-"
		}
		setCount := "-"
		if row.SetCount > 0 {
			setCount = strconv.Itoa(row.SetCount)
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			status,
			priority,
			dueDate,
			truncateString(row.BusinessUnit, buMax),
			truncateString(row.ServiceSite, siteMax),
			setCount,
		)
	}
	return writer.Flush()
}
