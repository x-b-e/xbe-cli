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

type workOrderAssignmentsListOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	NoAuth    bool
	Limit     int
	Offset    int
	Sort      string
	WorkOrder string
	User      string
}

type workOrderAssignmentRow struct {
	ID          string `json:"id"`
	WorkOrderID string `json:"work_order_id,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

func newWorkOrderAssignmentsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List work order assignments",
		Long: `List work order assignments.

Output Columns:
  ID          Assignment identifier
  WORK ORDER  Work order ID
  USER        Assigned user ID
  CREATED AT  Assignment creation timestamp

Filters:
  --work-order   Filter by work order ID
  --user         Filter by user ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List work order assignments
  xbe view work-order-assignments list

  # Filter by work order
  xbe view work-order-assignments list --work-order 123

  # Filter by user
  xbe view work-order-assignments list --user 456

  # Output as JSON
  xbe view work-order-assignments list --json`,
		Args: cobra.NoArgs,
		RunE: runWorkOrderAssignmentsList,
	}
	initWorkOrderAssignmentsListFlags(cmd)
	return cmd
}

func init() {
	workOrderAssignmentsCmd.AddCommand(newWorkOrderAssignmentsListCmd())
}

func initWorkOrderAssignmentsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("work-order", "", "Filter by work order ID")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runWorkOrderAssignmentsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseWorkOrderAssignmentsListOptions(cmd)
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
	query.Set("fields[work-order-assignments]", "created-at,updated-at,work-order,user")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[work-order]", opts.WorkOrder)
	setFilterIfPresent(query, "filter[user]", opts.User)

	body, _, err := client.Get(cmd.Context(), "/v1/work-order-assignments", query)
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

	rows := buildWorkOrderAssignmentRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderWorkOrderAssignmentsTable(cmd, rows)
}

func parseWorkOrderAssignmentsListOptions(cmd *cobra.Command) (workOrderAssignmentsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	workOrder, _ := cmd.Flags().GetString("work-order")
	user, _ := cmd.Flags().GetString("user")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return workOrderAssignmentsListOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		NoAuth:    noAuth,
		Limit:     limit,
		Offset:    offset,
		Sort:      sort,
		WorkOrder: workOrder,
		User:      user,
	}, nil
}

func buildWorkOrderAssignmentRows(resp jsonAPIResponse) []workOrderAssignmentRow {
	rows := make([]workOrderAssignmentRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := workOrderAssignmentRow{
			ID:          resource.ID,
			WorkOrderID: relationshipIDFromMap(resource.Relationships, "work-order"),
			UserID:      relationshipIDFromMap(resource.Relationships, "user"),
			CreatedAt:   formatDateTime(stringAttr(attrs, "created-at")),
			UpdatedAt:   formatDateTime(stringAttr(attrs, "updated-at")),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderWorkOrderAssignmentsTable(cmd *cobra.Command, rows []workOrderAssignmentRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No work order assignments found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tWORK_ORDER\tUSER\tCREATED_AT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			row.WorkOrderID,
			row.UserID,
			row.CreatedAt,
		)
	}
	return writer.Flush()
}
