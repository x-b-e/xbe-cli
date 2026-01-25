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

type geofenceRestrictionsListOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	NoAuth   bool
	Limit    int
	Offset   int
	Geofence string
	Trucker  string
	Status   string
}

type geofenceRestrictionRow struct {
	ID                            string `json:"id"`
	Status                        string `json:"status,omitempty"`
	MaxSecondsBetweenNotification string `json:"max_seconds_between_notification,omitempty"`
	GeofenceID                    string `json:"geofence_id,omitempty"`
	GeofenceName                  string `json:"geofence_name,omitempty"`
	TruckerID                     string `json:"trucker_id,omitempty"`
	TruckerName                   string `json:"trucker_name,omitempty"`
}

func newGeofenceRestrictionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List geofence restrictions",
		Long: `List geofence restrictions with filtering and pagination.

Geofence restrictions define custom trucker access rules for specific geofences.

Output Columns:
  ID               Restriction identifier
  GEOFENCE         Geofence name (or ID)
  TRUCKER          Trucker name (or ID)
  STATUS           Restriction status
  MAX SECONDS      Max seconds between notifications

Filters:
  --geofence        Filter by geofence ID
  --trucker         Filter by trucker ID
  --status          Filter by status`,
		Example: `  # List all geofence restrictions
  xbe view geofence-restrictions list

  # Filter by geofence
  xbe view geofence-restrictions list --geofence 123

  # Filter by trucker
  xbe view geofence-restrictions list --trucker 456

  # Filter by status
  xbe view geofence-restrictions list --status active

  # Output as JSON
  xbe view geofence-restrictions list --json`,
		RunE: runGeofenceRestrictionsList,
	}
	initGeofenceRestrictionsListFlags(cmd)
	return cmd
}

func init() {
	geofenceRestrictionsCmd.AddCommand(newGeofenceRestrictionsListCmd())
}

func initGeofenceRestrictionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("geofence", "", "Filter by geofence ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runGeofenceRestrictionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseGeofenceRestrictionsListOptions(cmd)
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
	query.Set("fields[geofence-restrictions]", "status,max-seconds-between-notification,geofence,trucker")
	query.Set("include", "geofence,trucker")
	query.Set("fields[geofences]", "name")
	query.Set("fields[truckers]", "company-name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[geofence]", opts.Geofence)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[status]", opts.Status)

	body, _, err := client.Get(cmd.Context(), "/v1/geofence-restrictions", query)
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

	rows := buildGeofenceRestrictionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderGeofenceRestrictionsTable(cmd, rows)
}

func parseGeofenceRestrictionsListOptions(cmd *cobra.Command) (geofenceRestrictionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	geofence, _ := cmd.Flags().GetString("geofence")
	trucker, _ := cmd.Flags().GetString("trucker")
	status, _ := cmd.Flags().GetString("status")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return geofenceRestrictionsListOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		NoAuth:   noAuth,
		Limit:    limit,
		Offset:   offset,
		Geofence: geofence,
		Trucker:  trucker,
		Status:   status,
	}, nil
}

func buildGeofenceRestrictionRows(resp jsonAPIResponse) []geofenceRestrictionRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]geofenceRestrictionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := geofenceRestrictionRow{
			ID:                            resource.ID,
			Status:                        stringAttr(resource.Attributes, "status"),
			MaxSecondsBetweenNotification: stringAttr(resource.Attributes, "max-seconds-between-notification"),
		}

		if rel, ok := resource.Relationships["geofence"]; ok && rel.Data != nil {
			row.GeofenceID = rel.Data.ID
			if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.GeofenceName = stringAttr(inc.Attributes, "name")
			}
		}

		if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
			row.TruckerID = rel.Data.ID
			if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.TruckerName = stringAttr(inc.Attributes, "company-name")
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderGeofenceRestrictionsTable(cmd *cobra.Command, rows []geofenceRestrictionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No geofence restrictions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tGEOFENCE\tTRUCKER\tSTATUS\tMAX SECONDS")
	for _, row := range rows {
		geofenceLabel := firstNonEmpty(row.GeofenceName, row.GeofenceID)
		truckerLabel := firstNonEmpty(row.TruckerName, row.TruckerID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(geofenceLabel, 25),
			truncateString(truckerLabel, 25),
			row.Status,
			row.MaxSecondsBetweenNotification,
		)
	}
	writer.Flush()
	return nil
}
