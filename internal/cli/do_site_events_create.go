package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doSiteEventsCreateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	EventType              string
	EventKind              string
	EventDetails           string
	EventAt                string
	EventLatitude          string
	EventLongitude         string
	TenderJobScheduleShift string
	MaterialTransaction    string
	EventSiteType          string
	EventSiteID            string
}

func newDoSiteEventsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a site event",
		Long: `Create a site event.

Required flags:
  --event-type             Event type (arrive_site,start_work,stop_work,depart_site,at_site)
  --event-at               Event timestamp (ISO 8601)
  --event-latitude         Event latitude
  --event-longitude        Event longitude
  --event-site-type        Event site type (job-sites, material-sites, parking-sites)
  --event-site-id          Event site ID
  --tender-job-schedule-shift or --material-transaction

Optional flags:
  --event-kind             Event kind (load,unload,pour; required for start_work/stop_work)
  --event-details          Event details`,
		Example: `  # Create a start-work site event
  xbe do site-events create \
    --event-type start_work \
    --event-kind load \
    --event-at 2025-01-01T12:00:00Z \
    --event-latitude 41.8781 \
    --event-longitude -87.6298 \
    --material-transaction 123 \
    --event-site-type material-sites \
    --event-site-id 456

  # Create an arrive-site event
  xbe do site-events create \
    --event-type arrive_site \
    --event-at 2025-01-01T12:00:00Z \
    --event-latitude 41.8781 \
    --event-longitude -87.6298 \
    --tender-job-schedule-shift 789 \
    --event-site-type job-sites \
    --event-site-id 321`,
		Args: cobra.NoArgs,
		RunE: runDoSiteEventsCreate,
	}
	initDoSiteEventsCreateFlags(cmd)
	return cmd
}

func init() {
	doSiteEventsCmd.AddCommand(newDoSiteEventsCreateCmd())
}

func initDoSiteEventsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("event-type", "", "Event type (arrive_site,start_work,stop_work,depart_site,at_site)")
	cmd.Flags().String("event-kind", "", "Event kind (load,unload,pour)")
	cmd.Flags().String("event-details", "", "Event details")
	cmd.Flags().String("event-at", "", "Event timestamp (ISO 8601)")
	cmd.Flags().String("event-latitude", "", "Event latitude")
	cmd.Flags().String("event-longitude", "", "Event longitude")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID")
	cmd.Flags().String("material-transaction", "", "Material transaction ID")
	cmd.Flags().String("event-site-type", "", "Event site type (job-sites, material-sites, parking-sites)")
	cmd.Flags().String("event-site-id", "", "Event site ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoSiteEventsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoSiteEventsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
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

	if opts.EventType == "" {
		err := fmt.Errorf("--event-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.EventAt == "" {
		err := fmt.Errorf("--event-at is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.EventLatitude == "" {
		err := fmt.Errorf("--event-latitude is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.EventLongitude == "" {
		err := fmt.Errorf("--event-longitude is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.EventSiteType == "" {
		err := fmt.Errorf("--event-site-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.EventSiteID == "" {
		err := fmt.Errorf("--event-site-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.TenderJobScheduleShift == "" && opts.MaterialTransaction == "" {
		err := fmt.Errorf("--tender-job-schedule-shift or --material-transaction is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if requiresEventKind(opts.EventType) && opts.EventKind == "" {
		err := fmt.Errorf("--event-kind is required when --event-type is %s", opts.EventType)
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if forbidsEventKind(opts.EventType) && opts.EventKind != "" {
		err := fmt.Errorf("--event-kind must be blank when --event-type is %s", opts.EventType)
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"event-type":      opts.EventType,
		"event-at":        opts.EventAt,
		"event-latitude":  opts.EventLatitude,
		"event-longitude": opts.EventLongitude,
	}
	if opts.EventKind != "" {
		attributes["event-kind"] = opts.EventKind
	}
	if opts.EventDetails != "" {
		attributes["event-details"] = opts.EventDetails
	}

	relationships := map[string]any{
		"event-site": map[string]any{
			"data": map[string]any{
				"type": opts.EventSiteType,
				"id":   opts.EventSiteID,
			},
		},
	}

	if opts.TenderJobScheduleShift != "" {
		relationships["tender-job-schedule-shift"] = map[string]any{
			"data": map[string]any{
				"type": "tender-job-schedule-shifts",
				"id":   opts.TenderJobScheduleShift,
			},
		}
	}
	if opts.MaterialTransaction != "" {
		relationships["material-transaction"] = map[string]any{
			"data": map[string]any{
				"type": "material-transactions",
				"id":   opts.MaterialTransaction,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "site-events",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/site-events", jsonBody)
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

	row := buildSiteEventRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created site event %s\n", row.ID)
	return nil
}

func parseDoSiteEventsCreateOptions(cmd *cobra.Command) (doSiteEventsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	eventType, _ := cmd.Flags().GetString("event-type")
	eventKind, _ := cmd.Flags().GetString("event-kind")
	eventDetails, _ := cmd.Flags().GetString("event-details")
	eventAt, _ := cmd.Flags().GetString("event-at")
	eventLatitude, _ := cmd.Flags().GetString("event-latitude")
	eventLongitude, _ := cmd.Flags().GetString("event-longitude")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	materialTransaction, _ := cmd.Flags().GetString("material-transaction")
	eventSiteType, _ := cmd.Flags().GetString("event-site-type")
	eventSiteID, _ := cmd.Flags().GetString("event-site-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doSiteEventsCreateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		EventType:              eventType,
		EventKind:              eventKind,
		EventDetails:           eventDetails,
		EventAt:                eventAt,
		EventLatitude:          eventLatitude,
		EventLongitude:         eventLongitude,
		TenderJobScheduleShift: tenderJobScheduleShift,
		MaterialTransaction:    materialTransaction,
		EventSiteType:          eventSiteType,
		EventSiteID:            eventSiteID,
	}, nil
}

func requiresEventKind(eventType string) bool {
	switch eventType {
	case "start_work", "stop_work":
		return true
	default:
		return false
	}
}

func forbidsEventKind(eventType string) bool {
	switch eventType {
	case "arrive_site", "depart_site":
		return true
	default:
		return false
	}
}
