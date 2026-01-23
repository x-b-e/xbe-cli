package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type transportRoutesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type transportRouteDetails struct {
	ID                   string   `json:"id"`
	OriginLatitude       string   `json:"origin_latitude,omitempty"`
	OriginLongitude      string   `json:"origin_longitude,omitempty"`
	DestinationLatitude  string   `json:"destination_latitude,omitempty"`
	DestinationLongitude string   `json:"destination_longitude,omitempty"`
	Miles                *float64 `json:"miles,omitempty"`
	Minutes              *float64 `json:"minutes,omitempty"`
	PolylineCoordinates  any      `json:"polyline_coordinates,omitempty"`
}

func newTransportRoutesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show transport route details",
		Long: `Show the full details of a transport route.

Output Fields:
  ID
  Origin Latitude
  Origin Longitude
  Destination Latitude
  Destination Longitude
  Miles
  Minutes
  Polyline Coordinates

Arguments:
  <id>    The transport route ID (required). You can find IDs using the list command.`,
		Example: `  # Show a route
  xbe view transport-routes show 123

  # Get JSON output
  xbe view transport-routes show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTransportRoutesShow,
	}
	initTransportRoutesShowFlags(cmd)
	return cmd
}

func init() {
	transportRoutesCmd.AddCommand(newTransportRoutesShowCmd())
}

func initTransportRoutesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTransportRoutesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTransportRoutesShowOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("transport route id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[transport-routes]", "origin-latitude,origin-longitude,destination-latitude,destination-longitude,miles,minutes,polyline-coordinates")

	body, _, err := client.Get(cmd.Context(), "/v1/transport-routes/"+id, query)
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

	details := buildTransportRouteDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTransportRouteDetails(cmd, details)
}

func parseTransportRoutesShowOptions(cmd *cobra.Command) (transportRoutesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return transportRoutesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTransportRouteDetails(resp jsonAPISingleResponse) transportRouteDetails {
	attrs := resp.Data.Attributes
	var milesPtr *float64
	var minutesPtr *float64

	if miles, ok := floatAttrValue(attrs, "miles"); ok {
		milesPtr = &miles
	}
	if minutes, ok := floatAttrValue(attrs, "minutes"); ok {
		minutesPtr = &minutes
	}

	return transportRouteDetails{
		ID:                   resp.Data.ID,
		OriginLatitude:       stringAttr(attrs, "origin-latitude"),
		OriginLongitude:      stringAttr(attrs, "origin-longitude"),
		DestinationLatitude:  stringAttr(attrs, "destination-latitude"),
		DestinationLongitude: stringAttr(attrs, "destination-longitude"),
		Miles:                milesPtr,
		Minutes:              minutesPtr,
		PolylineCoordinates:  attrs["polyline-coordinates"],
	}
}

func renderTransportRouteDetails(cmd *cobra.Command, details transportRouteDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.OriginLatitude != "" {
		fmt.Fprintf(out, "Origin Latitude: %s\n", details.OriginLatitude)
	}
	if details.OriginLongitude != "" {
		fmt.Fprintf(out, "Origin Longitude: %s\n", details.OriginLongitude)
	}
	if details.DestinationLatitude != "" {
		fmt.Fprintf(out, "Destination Latitude: %s\n", details.DestinationLatitude)
	}
	if details.DestinationLongitude != "" {
		fmt.Fprintf(out, "Destination Longitude: %s\n", details.DestinationLongitude)
	}
	if details.Miles != nil {
		fmt.Fprintf(out, "Miles: %s\n", formatMiles(*details.Miles))
	}
	if details.Minutes != nil {
		fmt.Fprintf(out, "Minutes: %s\n", formatMinutes(*details.Minutes))
	}
	if details.PolylineCoordinates != nil {
		pretty, err := json.MarshalIndent(details.PolylineCoordinates, "", "  ")
		if err != nil {
			fmt.Fprintf(out, "Polyline Coordinates: %v\n", details.PolylineCoordinates)
			return nil
		}
		fmt.Fprintln(out, "Polyline Coordinates:")
		fmt.Fprintln(out, string(pretty))
	}

	return nil
}

func floatAttrValue(attrs map[string]any, key string) (float64, bool) {
	if attrs == nil {
		return 0, false
	}
	value, ok := attrs[key]
	if !ok || value == nil {
		return 0, false
	}
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case string:
		if f, err := strconv.ParseFloat(strings.TrimSpace(typed), 64); err == nil {
			return f, true
		}
	}
	return 0, false
}
