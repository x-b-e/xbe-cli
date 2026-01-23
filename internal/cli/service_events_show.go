package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type serviceEventsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type serviceEventDetails struct {
	ID                               string `json:"id"`
	TenderJobScheduleShiftID         string `json:"tender_job_schedule_shift_id,omitempty"`
	OccurredAt                       string `json:"occurred_at,omitempty"`
	Kind                             string `json:"kind,omitempty"`
	Note                             string `json:"note,omitempty"`
	OccurredLatitude                 string `json:"occurred_latitude,omitempty"`
	OccurredLongitude                string `json:"occurred_longitude,omitempty"`
	ViaGPS                           bool   `json:"via_gps"`
	ViaMaterialTransactionAcceptance bool   `json:"via_material_transaction_acceptance"`
	MilesToStartSite                 string `json:"miles_to_start_site,omitempty"`
	CreatedByID                      string `json:"created_by_id,omitempty"`
	UpdatedByID                      string `json:"updated_by_id,omitempty"`
	CreatedAt                        string `json:"created_at,omitempty"`
	UpdatedAt                        string `json:"updated_at,omitempty"`
}

func newServiceEventsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show service event details",
		Long: `Show the full details of a service event.

Output Fields:
  ID                Service event identifier
  Shift             Tender job schedule shift ID
  Kind              Event kind
  Occurred At       Event timestamp
  Note              Event note
  Occurred Latitude Event latitude
  Occurred Longitude Event longitude
  Via GPS           Whether event was via GPS
  Via Material Txn  Whether event was via material transaction acceptance
  Miles To Start    Distance to the shift start site (miles)
  Created By        User who created the event
  Updated By        User who last updated the event
  Created At        Created timestamp
  Updated At        Updated timestamp

Arguments:
  <id>    The service event ID (required). You can find IDs using the list command.`,
		Example: `  # Show a service event
  xbe view service-events show 123

  # Get JSON output
  xbe view service-events show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runServiceEventsShow,
	}
	initServiceEventsShowFlags(cmd)
	return cmd
}

func init() {
	serviceEventsCmd.AddCommand(newServiceEventsShowCmd())
}

func initServiceEventsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runServiceEventsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseServiceEventsShowOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("service event id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[service-events]", "occurred-at,kind,note,occurred-latitude,occurred-longitude,via-gps,via-material-transaction-acceptance,miles-to-start-site,created-at,updated-at,tender-job-schedule-shift,created-by,updated-by")

	body, _, err := client.Get(cmd.Context(), "/v1/service-events/"+id, query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	details := buildServiceEventDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderServiceEventDetails(cmd, details)
}

func parseServiceEventsShowOptions(cmd *cobra.Command) (serviceEventsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return serviceEventsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildServiceEventDetails(resp jsonAPISingleResponse) serviceEventDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := serviceEventDetails{
		ID:                               resource.ID,
		OccurredAt:                       formatDateTime(stringAttr(attrs, "occurred-at")),
		Kind:                             stringAttr(attrs, "kind"),
		Note:                             stringAttr(attrs, "note"),
		OccurredLatitude:                 stringAttr(attrs, "occurred-latitude"),
		OccurredLongitude:                stringAttr(attrs, "occurred-longitude"),
		ViaGPS:                           boolAttr(attrs, "via-gps"),
		ViaMaterialTransactionAcceptance: boolAttr(attrs, "via-material-transaction-acceptance"),
		CreatedAt:                        formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:                        formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if raw, ok := attrs["miles-to-start-site"]; ok && raw != nil {
		switch typed := raw.(type) {
		case float64:
			details.MilesToStartSite = fmt.Sprintf("%.2f", typed)
		case float32:
			details.MilesToStartSite = fmt.Sprintf("%.2f", typed)
		case int:
			details.MilesToStartSite = fmt.Sprintf("%d", typed)
		case int64:
			details.MilesToStartSite = fmt.Sprintf("%d", typed)
		case string:
			details.MilesToStartSite = typed
		default:
			details.MilesToStartSite = fmt.Sprintf("%v", typed)
		}
	}

	if rel, ok := resource.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		details.TenderJobScheduleShiftID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["updated-by"]; ok && rel.Data != nil {
		details.UpdatedByID = rel.Data.ID
	}

	return details
}

func renderServiceEventDetails(cmd *cobra.Command, details serviceEventDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TenderJobScheduleShiftID != "" {
		fmt.Fprintf(out, "Shift: %s\n", details.TenderJobScheduleShiftID)
	}
	if details.Kind != "" {
		fmt.Fprintf(out, "Kind: %s\n", details.Kind)
	}
	if details.OccurredAt != "" {
		fmt.Fprintf(out, "Occurred At: %s\n", details.OccurredAt)
	}
	if details.Note != "" {
		fmt.Fprintf(out, "Note: %s\n", details.Note)
	}
	if details.OccurredLatitude != "" {
		fmt.Fprintf(out, "Occurred Latitude: %s\n", details.OccurredLatitude)
	}
	if details.OccurredLongitude != "" {
		fmt.Fprintf(out, "Occurred Longitude: %s\n", details.OccurredLongitude)
	}
	fmt.Fprintf(out, "Via GPS: %t\n", details.ViaGPS)
	fmt.Fprintf(out, "Via Material Txn: %t\n", details.ViaMaterialTransactionAcceptance)
	if details.MilesToStartSite != "" {
		fmt.Fprintf(out, "Miles To Start: %s\n", details.MilesToStartSite)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	if details.UpdatedByID != "" {
		fmt.Fprintf(out, "Updated By: %s\n", details.UpdatedByID)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
