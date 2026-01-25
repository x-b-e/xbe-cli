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

type customerSettingsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
}

type customerSettingRow struct {
	ID                              string `json:"id"`
	CustomerID                      string `json:"customer_id,omitempty"`
	IsAuditingTimeCardApprovals     bool   `json:"is_auditing_time_card_approvals"`
	EnableRecapNotifications        bool   `json:"enable_recap_notifications"`
	PlanRequiresProject             bool   `json:"plan_requires_project"`
	PlanRequiresBusinessUnit        bool   `json:"plan_requires_business_unit"`
	AutoCancelShiftsWithoutActivity bool   `json:"auto_cancel_shifts_without_activity"`
}

func newCustomerSettingsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List customer settings",
		Long: `List customer settings.

Output Columns:
  ID                     Setting identifier
  CUSTOMER ID            Associated customer ID
  AUDIT TC               Auditing time card approvals
  RECAP NOTIF            Enable recap notifications
  REQ PROJECT            Plan requires project

Use --json for full setting details.`,
		Example: `  # List all customer settings
  xbe view customer-settings list

  # Output as JSON for full details
  xbe view customer-settings list --json`,
		RunE: runCustomerSettingsList,
	}
	initCustomerSettingsListFlags(cmd)
	return cmd
}

func init() {
	customerSettingsCmd.AddCommand(newCustomerSettingsListCmd())
}

func initCustomerSettingsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCustomerSettingsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCustomerSettingsListOptions(cmd)
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
	query.Set("include", "customer")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	body, _, err := client.Get(cmd.Context(), "/v1/customer-settings", query)
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

	rows := buildCustomerSettingRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCustomerSettingsTable(cmd, rows)
}

func parseCustomerSettingsListOptions(cmd *cobra.Command) (customerSettingsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return customerSettingsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
	}, nil
}

func buildCustomerSettingRows(resp jsonAPIResponse) []customerSettingRow {
	rows := make([]customerSettingRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := customerSettingRow{
			ID:                              resource.ID,
			IsAuditingTimeCardApprovals:     boolAttr(resource.Attributes, "is-auditing-time-card-approvals"),
			EnableRecapNotifications:        boolAttr(resource.Attributes, "enable-recap-notifications"),
			PlanRequiresProject:             boolAttr(resource.Attributes, "plan-requires-project"),
			PlanRequiresBusinessUnit:        boolAttr(resource.Attributes, "plan-requires-business-unit"),
			AutoCancelShiftsWithoutActivity: boolAttr(resource.Attributes, "auto-cancel-shifts-without-activity"),
		}

		if rel, ok := resource.Relationships["customer"]; ok && rel.Data != nil {
			row.CustomerID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderCustomerSettingsTable(cmd *cobra.Command, rows []customerSettingRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No customer settings found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tCUSTOMER ID\tAUDIT TC\tRECAP NOTIF\tREQ PROJECT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.CustomerID,
			boolToYesNo(row.IsAuditingTimeCardApprovals),
			boolToYesNo(row.EnableRecapNotifications),
			boolToYesNo(row.PlanRequiresProject),
		)
	}
	return writer.Flush()
}
