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

type doSiteEventsUpdateOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	ID             string
	EventDetails   string
	EventAt        string
	EventLatitude  string
	EventLongitude string
}

func newDoSiteEventsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a site event",
		Long: `Update a site event.

Optional attributes:
  --event-details    Event details
  --event-at         Event timestamp (ISO 8601)
  --event-latitude   Event latitude
  --event-longitude  Event longitude`,
		Example: `  # Update event details and coordinates
  xbe do site-events update 123 \
    --event-details "Updated notes" \
    --event-latitude 41.8781 \
    --event-longitude -87.6298`,
		Args: cobra.ExactArgs(1),
		RunE: runDoSiteEventsUpdate,
	}
	initDoSiteEventsUpdateFlags(cmd)
	return cmd
}

func init() {
	doSiteEventsCmd.AddCommand(newDoSiteEventsUpdateCmd())
}

func initDoSiteEventsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("event-details", "", "Event details")
	cmd.Flags().String("event-at", "", "Event timestamp (ISO 8601)")
	cmd.Flags().String("event-latitude", "", "Event latitude")
	cmd.Flags().String("event-longitude", "", "Event longitude")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoSiteEventsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoSiteEventsUpdateOptions(cmd, args)
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

	attributes := map[string]any{}
	if cmd.Flags().Changed("event-details") {
		attributes["event-details"] = opts.EventDetails
	}
	if cmd.Flags().Changed("event-at") {
		attributes["event-at"] = opts.EventAt
	}
	if cmd.Flags().Changed("event-latitude") {
		attributes["event-latitude"] = opts.EventLatitude
	}
	if cmd.Flags().Changed("event-longitude") {
		attributes["event-longitude"] = opts.EventLongitude
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "site-events",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/site-events/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated site event %s\n", row.ID)
	return nil
}

func parseDoSiteEventsUpdateOptions(cmd *cobra.Command, args []string) (doSiteEventsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	eventDetails, _ := cmd.Flags().GetString("event-details")
	eventAt, _ := cmd.Flags().GetString("event-at")
	eventLatitude, _ := cmd.Flags().GetString("event-latitude")
	eventLongitude, _ := cmd.Flags().GetString("event-longitude")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doSiteEventsUpdateOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		ID:             args[0],
		EventDetails:   eventDetails,
		EventAt:        eventAt,
		EventLatitude:  eventLatitude,
		EventLongitude: eventLongitude,
	}, nil
}
