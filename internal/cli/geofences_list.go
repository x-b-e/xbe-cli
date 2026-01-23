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

type geofencesListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Broker  string
	Name    string
}

type geofenceRow struct {
	ID              string `json:"id"`
	Name            string `json:"name,omitempty"`
	Description     string `json:"description,omitempty"`
	Status          string `json:"status,omitempty"`
	RestrictionMode string `json:"restriction_mode,omitempty"`
	BrokerID        string `json:"broker_id,omitempty"`
}

func newGeofencesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List geofences",
		Long: `List geofences (geographic boundaries).

Output Columns:
  ID               Geofence identifier
  NAME             Geofence name
  DESCRIPTION      Geofence description
  STATUS           Geofence status
  RESTRICTION      Restriction mode
  BROKER ID        Associated broker ID

Filters:
  --broker         Filter by broker ID
  --name           Filter by name`,
		Example: `  # List all geofences
  xbe view geofences list

  # Filter by broker
  xbe view geofences list --broker 123

  # Filter by name
  xbe view geofences list --name "Office"

  # Output as JSON
  xbe view geofences list --json`,
		RunE: runGeofencesList,
	}
	initGeofencesListFlags(cmd)
	return cmd
}

func init() {
	geofencesCmd.AddCommand(newGeofencesListCmd())
}

func initGeofencesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("name", "", "Filter by name")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runGeofencesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseGeofencesListOptions(cmd)
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

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[name]", opts.Name)

	body, _, err := client.Get(cmd.Context(), "/v1/geofences", query)
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

	rows := buildGeofenceRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderGeofencesTable(cmd, rows)
}

func parseGeofencesListOptions(cmd *cobra.Command) (geofencesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	broker, _ := cmd.Flags().GetString("broker")
	name, _ := cmd.Flags().GetString("name")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return geofencesListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Broker:  broker,
		Name:    name,
	}, nil
}

func buildGeofenceRows(resp jsonAPIResponse) []geofenceRow {
	rows := make([]geofenceRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := geofenceRow{
			ID:              resource.ID,
			Name:            stringAttr(resource.Attributes, "name"),
			Description:     stringAttr(resource.Attributes, "description"),
			Status:          stringAttr(resource.Attributes, "status"),
			RestrictionMode: stringAttr(resource.Attributes, "restriction-mode"),
		}

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderGeofencesTable(cmd *cobra.Command, rows []geofenceRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No geofences found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tDESCRIPTION\tSTATUS\tRESTRICTION\tBROKER ID")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 25),
			truncateString(row.Description, 30),
			row.Status,
			row.RestrictionMode,
			row.BrokerID,
		)
	}
	return writer.Flush()
}
