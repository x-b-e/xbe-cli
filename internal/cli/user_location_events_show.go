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

type userLocationEventsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type userLocationEventDetails struct {
	ID             string `json:"id"`
	EventAt        string `json:"event_at,omitempty"`
	EventLatitude  string `json:"event_latitude,omitempty"`
	EventLongitude string `json:"event_longitude,omitempty"`
	Provenance     string `json:"provenance,omitempty"`
	UserID         string `json:"user_id,omitempty"`
	UpdatedByID    string `json:"updated_by_id,omitempty"`
}

func newUserLocationEventsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show user location event details",
		Long: `Show the full details of a user location event.

Output Fields:
  ID          User location event identifier
  Event At    Event timestamp
  Latitude    Event latitude
  Longitude   Event longitude
  Provenance  Event provenance
  User        User ID
  Updated By  Updated-by user ID

Arguments:
  <id>   The user location event ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a user location event
  xbe view user-location-events show 123

  # JSON output
  xbe view user-location-events show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runUserLocationEventsShow,
	}
	initUserLocationEventsShowFlags(cmd)
	return cmd
}

func init() {
	userLocationEventsCmd.AddCommand(newUserLocationEventsShowCmd())
}

func initUserLocationEventsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runUserLocationEventsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseUserLocationEventsShowOptions(cmd)
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
		return fmt.Errorf("user location event id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[user-location-events]", "event-at,event-latitude,event-longitude,provenance,user,updated-by")

	body, _, err := client.Get(cmd.Context(), "/v1/user-location-events/"+id, query)
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

	details := buildUserLocationEventDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderUserLocationEventDetails(cmd, details)
}

func parseUserLocationEventsShowOptions(cmd *cobra.Command) (userLocationEventsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return userLocationEventsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildUserLocationEventDetails(resp jsonAPISingleResponse) userLocationEventDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := userLocationEventDetails{
		ID:             resource.ID,
		EventAt:        formatDateTime(stringAttr(attrs, "event-at")),
		EventLatitude:  stringAttr(attrs, "event-latitude"),
		EventLongitude: stringAttr(attrs, "event-longitude"),
		Provenance:     stringAttr(attrs, "provenance"),
	}

	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["updated-by"]; ok && rel.Data != nil {
		details.UpdatedByID = rel.Data.ID
	}

	return details
}

func renderUserLocationEventDetails(cmd *cobra.Command, details userLocationEventDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.EventAt != "" {
		fmt.Fprintf(out, "Event At: %s\n", details.EventAt)
	}
	if details.EventLatitude != "" {
		fmt.Fprintf(out, "Latitude: %s\n", details.EventLatitude)
	}
	if details.EventLongitude != "" {
		fmt.Fprintf(out, "Longitude: %s\n", details.EventLongitude)
	}
	if details.Provenance != "" {
		fmt.Fprintf(out, "Provenance: %s\n", details.Provenance)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User: %s\n", details.UserID)
	}
	if details.UpdatedByID != "" {
		fmt.Fprintf(out, "Updated By: %s\n", details.UpdatedByID)
	}

	return nil
}
