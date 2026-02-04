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

type placesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type placeDetails struct {
	ID               string   `json:"id"`
	PlaceID          string   `json:"place_id,omitempty"`
	FormattedAddress string   `json:"formatted_address,omitempty"`
	Latitude         *float64 `json:"latitude,omitempty"`
	Longitude        *float64 `json:"longitude,omitempty"`
}

func newPlacesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <place-id>",
		Short: "Show place details",
		Long: `Show the full details of a Google Place.

Output Fields:
  ID
  Place ID
  Formatted Address
  Latitude
  Longitude

Arguments:
  <place-id>  The Google Place ID (required). You can obtain this from
              the places predictions API or other Google Places services.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a place by ID
  xbe view places show ChIJD7fiBh9u5kcRYJSMaMOCCwQ

  # Get JSON output
  xbe view places show ChIJD7fiBh9u5kcRYJSMaMOCCwQ --json`,
		Args: cobra.ExactArgs(1),
		RunE: runPlacesShow,
	}
	initPlacesShowFlags(cmd)
	return cmd
}

func init() {
	placesCmd.AddCommand(newPlacesShowCmd())
}

func initPlacesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPlacesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parsePlacesShowOptions(cmd)
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

	placeID := strings.TrimSpace(args[0])
	if placeID == "" {
		return fmt.Errorf("place id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/places/"+placeID, nil)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildPlaceDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderPlaceDetails(cmd, details)
}

func parsePlacesShowOptions(cmd *cobra.Command) (placesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return placesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildPlaceDetails(resp jsonAPISingleResponse) placeDetails {
	resource := resp.Data
	attrs := resource.Attributes
	return placeDetails{
		ID:               resource.ID,
		PlaceID:          stringAttr(attrs, "place-id"),
		FormattedAddress: stringAttr(attrs, "formatted-address"),
		Latitude:         floatAttrPointer(attrs, "latitude"),
		Longitude:        floatAttrPointer(attrs, "longitude"),
	}
}

func renderPlaceDetails(cmd *cobra.Command, details placeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.PlaceID != "" {
		fmt.Fprintf(out, "Place ID: %s\n", details.PlaceID)
	}
	if details.FormattedAddress != "" {
		fmt.Fprintf(out, "Formatted Address: %s\n", details.FormattedAddress)
	}
	if details.Latitude != nil {
		fmt.Fprintf(out, "Latitude: %.6f\n", *details.Latitude)
	}
	if details.Longitude != nil {
		fmt.Fprintf(out, "Longitude: %.6f\n", *details.Longitude)
	}

	return nil
}
