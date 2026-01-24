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

type doUserLocationEventsCreateOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	User           string
	EventAt        string
	EventLatitude  string
	EventLongitude string
	Provenance     string
}

func newDoUserLocationEventsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a user location event",
		Long: `Create a user location event.

Required flags:
  --user        User ID for the event
  --provenance  Event provenance (gps,map)

Optional flags:
  --event-at         Event timestamp (ISO 8601; defaults to now)
  --event-latitude   Event latitude
  --event-longitude  Event longitude

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a user location event
  xbe do user-location-events create --user 123 --provenance gps \
    --event-at 2025-01-01T12:00:00Z --event-latitude 40.7128 --event-longitude -74.0060

  # Create using server timestamp
  xbe do user-location-events create --user 123 --provenance map --event-latitude 41.8781 --event-longitude -87.6298`,
		Args: cobra.NoArgs,
		RunE: runDoUserLocationEventsCreate,
	}
	initDoUserLocationEventsCreateFlags(cmd)
	return cmd
}

func init() {
	doUserLocationEventsCmd.AddCommand(newDoUserLocationEventsCreateCmd())
}

func initDoUserLocationEventsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("user", "", "User ID (required)")
	cmd.Flags().String("event-at", "", "Event timestamp (ISO 8601)")
	cmd.Flags().String("event-latitude", "", "Event latitude")
	cmd.Flags().String("event-longitude", "", "Event longitude")
	cmd.Flags().String("provenance", "", "Event provenance (gps,map)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("user")
	_ = cmd.MarkFlagRequired("provenance")
}

func runDoUserLocationEventsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoUserLocationEventsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.User) == "" {
		err := fmt.Errorf("--user is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Provenance) == "" {
		err := fmt.Errorf("--provenance is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"provenance": opts.Provenance,
	}
	if strings.TrimSpace(opts.EventAt) != "" {
		attributes["event-at"] = opts.EventAt
	}
	if strings.TrimSpace(opts.EventLatitude) != "" {
		attributes["event-latitude"] = opts.EventLatitude
	}
	if strings.TrimSpace(opts.EventLongitude) != "" {
		attributes["event-longitude"] = opts.EventLongitude
	}

	relationships := map[string]any{
		"user": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.User,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "user-location-events",
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

	body, _, err := client.Post(cmd.Context(), "/v1/user-location-events", jsonBody)
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

	row := buildUserLocationEventRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created user location event %s\n", row.ID)
	return nil
}

func parseDoUserLocationEventsCreateOptions(cmd *cobra.Command) (doUserLocationEventsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	user, _ := cmd.Flags().GetString("user")
	eventAt, _ := cmd.Flags().GetString("event-at")
	eventLatitude, _ := cmd.Flags().GetString("event-latitude")
	eventLongitude, _ := cmd.Flags().GetString("event-longitude")
	provenance, _ := cmd.Flags().GetString("provenance")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doUserLocationEventsCreateOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		User:           user,
		EventAt:        eventAt,
		EventLatitude:  eventLatitude,
		EventLongitude: eventLongitude,
		Provenance:     provenance,
	}, nil
}
