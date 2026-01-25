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

type truckerSettingsListOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	NoAuth    bool
	Limit     int
	Offset    int
	Trucker   string
	TruckerID string
}

type truckerSettingRow struct {
	ID                                                    string `json:"id"`
	TruckerID                                             string `json:"trucker_id,omitempty"`
	NotifyDriverWhenGPSNotAvailable                       bool   `json:"notify_driver_when_gps_not_available"`
	MinimumDriverTrackingMinutes                          int    `json:"minimum_driver_tracking_minutes"`
	AutoGenerateTimeSheetLineItemsPerJob                  bool   `json:"auto_generate_time_sheet_line_items_per_job"`
	RestrictLineItemClassificationEditToTimeSheetApprover bool   `json:"restrict_line_item_classification_edit_to_time_sheet_approver"`
	AutoCombineOverlappingDriverDays                      bool   `json:"auto_combine_overlapping_driver_days"`
}

func newTruckerSettingsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List trucker settings",
		Long: `List trucker settings.

Output Columns:
  ID                     Setting identifier
  TRUCKER ID             Associated trucker ID
  GPS NOTIF              Notify driver when GPS is unavailable
  MIN TRACK MINS         Minimum driver tracking minutes
  AUTO TS ITEMS          Auto-generate time sheet line items per job
  RESTRICT LI EDIT       Restrict line item edits to time sheet approver
  AUTO COMBINE DAYS      Auto-combine overlapping driver days

Filters:
  --trucker     Filter by trucker ID
  --trucker-id  Filter by trucker ID (alias of --trucker)

Use --json for machine-readable output or run 'xbe view trucker-settings show' for full details.`,
		Example: `  # List all trucker settings
  xbe view trucker-settings list

  # Filter by trucker
  xbe view trucker-settings list --trucker 123

  # Output as JSON
  xbe view trucker-settings list --json`,
		RunE: runTruckerSettingsList,
	}
	initTruckerSettingsListFlags(cmd)
	return cmd
}

func init() {
	truckerSettingsCmd.AddCommand(newTruckerSettingsListCmd())
}

func initTruckerSettingsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("trucker-id", "", "Filter by trucker ID (alias of --trucker)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTruckerSettingsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTruckerSettingsListOptions(cmd)
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
	query.Set("include", "trucker")
	query.Set("fields[trucker-settings]", "trucker,notify-driver-when-gps-not-available,minimum-driver-tracking-minutes,auto-generate-time-sheet-line-items-per-job,restrict-line-item-classification-edit-to-time-sheet-approver,auto-combine-overlapping-driver-days")
	query.Set("fields[truckers]", "company-name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[trucker_id]", opts.TruckerID)

	body, _, err := client.Get(cmd.Context(), "/v1/trucker-settings", query)
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

	rows := buildTruckerSettingRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTruckerSettingsTable(cmd, rows)
}

func parseTruckerSettingsListOptions(cmd *cobra.Command) (truckerSettingsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	trucker, _ := cmd.Flags().GetString("trucker")
	truckerID, _ := cmd.Flags().GetString("trucker-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return truckerSettingsListOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		NoAuth:    noAuth,
		Limit:     limit,
		Offset:    offset,
		Trucker:   trucker,
		TruckerID: truckerID,
	}, nil
}

func buildTruckerSettingRows(resp jsonAPIResponse) []truckerSettingRow {
	rows := make([]truckerSettingRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := truckerSettingRow{
			ID:                                   resource.ID,
			NotifyDriverWhenGPSNotAvailable:      boolAttr(resource.Attributes, "notify-driver-when-gps-not-available"),
			MinimumDriverTrackingMinutes:         intAttr(resource.Attributes, "minimum-driver-tracking-minutes"),
			AutoGenerateTimeSheetLineItemsPerJob: boolAttr(resource.Attributes, "auto-generate-time-sheet-line-items-per-job"),
			RestrictLineItemClassificationEditToTimeSheetApprover: boolAttr(resource.Attributes, "restrict-line-item-classification-edit-to-time-sheet-approver"),
			AutoCombineOverlappingDriverDays:                      boolAttr(resource.Attributes, "auto-combine-overlapping-driver-days"),
		}

		if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
			row.TruckerID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderTruckerSettingsTable(cmd *cobra.Command, rows []truckerSettingRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No trucker settings found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTRUCKER ID\tGPS NOTIF\tMIN TRACK MINS\tAUTO TS ITEMS\tRESTRICT LI EDIT\tAUTO COMBINE DAYS")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%d\t%s\t%s\t%s\n",
			row.ID,
			row.TruckerID,
			boolToYesNo(row.NotifyDriverWhenGPSNotAvailable),
			row.MinimumDriverTrackingMinutes,
			boolToYesNo(row.AutoGenerateTimeSheetLineItemsPerJob),
			boolToYesNo(row.RestrictLineItemClassificationEditToTimeSheetApprover),
			boolToYesNo(row.AutoCombineOverlappingDriverDays),
		)
	}
	return writer.Flush()
}
