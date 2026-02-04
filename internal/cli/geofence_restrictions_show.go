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

type geofenceRestrictionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type geofenceRestrictionDetails struct {
	ID                            string `json:"id"`
	Status                        string `json:"status"`
	MaxSecondsBetweenNotification string `json:"max_seconds_between_notification,omitempty"`
	GeofenceID                    string `json:"geofence_id,omitempty"`
	GeofenceName                  string `json:"geofence_name,omitempty"`
	TruckerID                     string `json:"trucker_id,omitempty"`
	TruckerName                   string `json:"trucker_name,omitempty"`
}

func newGeofenceRestrictionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show geofence restriction details",
		Long: `Show the full details of a specific geofence restriction.

Geofence restrictions define custom trucker access rules for specific geofences.

Output Fields:
  ID               Restriction identifier
  Status           Restriction status
  Max Seconds      Max seconds between notifications
  Geofence         Geofence name (or ID)
  Trucker          Trucker name (or ID)

Arguments:
  <id>             The geofence restriction ID (required).`,
		Example: `  # Show a geofence restriction
  xbe view geofence-restrictions show 123

  # Show as JSON
  xbe view geofence-restrictions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runGeofenceRestrictionsShow,
	}
	initGeofenceRestrictionsShowFlags(cmd)
	return cmd
}

func init() {
	geofenceRestrictionsCmd.AddCommand(newGeofenceRestrictionsShowCmd())
}

func initGeofenceRestrictionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runGeofenceRestrictionsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseGeofenceRestrictionsShowOptions(cmd)
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
		return fmt.Errorf("geofence restriction id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[geofence-restrictions]", "status,max-seconds-between-notification,geofence,trucker")
	query.Set("include", "geofence,trucker")
	query.Set("fields[geofences]", "name")
	query.Set("fields[truckers]", "company-name")

	body, _, err := client.Get(cmd.Context(), "/v1/geofence-restrictions/"+id, query)
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

	details := buildGeofenceRestrictionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderGeofenceRestrictionDetails(cmd, details)
}

func parseGeofenceRestrictionsShowOptions(cmd *cobra.Command) (geofenceRestrictionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return geofenceRestrictionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildGeofenceRestrictionDetails(resp jsonAPISingleResponse) geofenceRestrictionDetails {
	attrs := resp.Data.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := geofenceRestrictionDetails{
		ID:                            resp.Data.ID,
		Status:                        stringAttr(attrs, "status"),
		MaxSecondsBetweenNotification: stringAttr(attrs, "max-seconds-between-notification"),
	}

	if rel, ok := resp.Data.Relationships["geofence"]; ok && rel.Data != nil {
		details.GeofenceID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.GeofenceName = stringAttr(inc.Attributes, "name")
		}
	}

	if rel, ok := resp.Data.Relationships["trucker"]; ok && rel.Data != nil {
		details.TruckerID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.TruckerName = stringAttr(inc.Attributes, "company-name")
		}
	}

	return details
}

func renderGeofenceRestrictionDetails(cmd *cobra.Command, details geofenceRestrictionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.MaxSecondsBetweenNotification != "" {
		fmt.Fprintf(out, "Max Seconds Between Notifications: %s\n", details.MaxSecondsBetweenNotification)
	}

	geofenceLabel := firstNonEmpty(details.GeofenceName, details.GeofenceID)
	if geofenceLabel != "" {
		fmt.Fprintf(out, "Geofence: %s\n", geofenceLabel)
	}

	truckerLabel := firstNonEmpty(details.TruckerName, details.TruckerID)
	if truckerLabel != "" {
		fmt.Fprintf(out, "Trucker: %s\n", truckerLabel)
	}

	return nil
}
