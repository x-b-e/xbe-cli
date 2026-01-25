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

type brokerSettingsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
}

type brokerSettingRow struct {
	ID                                   string `json:"id"`
	BrokerID                             string `json:"broker_id,omitempty"`
	IsAuditingTimeCardApprovals          bool   `json:"is_auditing_time_card_approvals"`
	EnableRecapNotifications             bool   `json:"enable_recap_notifications"`
	PlanRequiresProject                  bool   `json:"plan_requires_project"`
	PlanRequiresBusinessUnit             bool   `json:"plan_requires_business_unit"`
	AutoCancelShiftsWithoutActivity      bool   `json:"auto_cancel_shifts_without_activity"`
	RestrictContactInfoVisibility        bool   `json:"restrict_contact_info_visibility"`
	RequireExplicitRateEditingPermission bool   `json:"require_explicit_rate_editing_permission"`
}

func newBrokerSettingsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List broker settings",
		Long: `List broker settings.

Output Columns:
  ID                     Setting identifier
  BROKER ID              Associated broker ID
  AUDIT TC               Auditing time card approvals
  RECAP NOTIF            Enable recap notifications
  REQ PROJECT            Plan requires project

Use --json for full setting details.`,
		Example: `  # List all broker settings
  xbe view broker-settings list

  # Output as JSON for full details
  xbe view broker-settings list --json`,
		RunE: runBrokerSettingsList,
	}
	initBrokerSettingsListFlags(cmd)
	return cmd
}

func init() {
	brokerSettingsCmd.AddCommand(newBrokerSettingsListCmd())
}

func initBrokerSettingsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBrokerSettingsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseBrokerSettingsListOptions(cmd)
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
	query.Set("include", "broker")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	body, _, err := client.Get(cmd.Context(), "/v1/broker-settings", query)
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

	rows := buildBrokerSettingRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderBrokerSettingsTable(cmd, rows)
}

func parseBrokerSettingsListOptions(cmd *cobra.Command) (brokerSettingsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return brokerSettingsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
	}, nil
}

func buildBrokerSettingRows(resp jsonAPIResponse) []brokerSettingRow {
	rows := make([]brokerSettingRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := brokerSettingRow{
			ID:                                   resource.ID,
			IsAuditingTimeCardApprovals:          boolAttr(resource.Attributes, "is-auditing-time-card-approvals"),
			EnableRecapNotifications:             boolAttr(resource.Attributes, "enable-recap-notifications"),
			PlanRequiresProject:                  boolAttr(resource.Attributes, "plan-requires-project"),
			PlanRequiresBusinessUnit:             boolAttr(resource.Attributes, "plan-requires-business-unit"),
			AutoCancelShiftsWithoutActivity:      boolAttr(resource.Attributes, "auto-cancel-shifts-without-activity"),
			RestrictContactInfoVisibility:        boolAttr(resource.Attributes, "restrict-contact-info-visibility"),
			RequireExplicitRateEditingPermission: boolAttr(resource.Attributes, "require-explicit-rate-editing-permission"),
		}

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderBrokerSettingsTable(cmd *cobra.Command, rows []brokerSettingRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No broker settings found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tBROKER ID\tAUDIT TC\tRECAP NOTIF\tREQ PROJECT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.BrokerID,
			boolToYesNo(row.IsAuditingTimeCardApprovals),
			boolToYesNo(row.EnableRecapNotifications),
			boolToYesNo(row.PlanRequiresProject),
		)
	}
	return writer.Flush()
}

func boolToYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
