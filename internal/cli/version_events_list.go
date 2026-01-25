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

type versionEventsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	CreatedAtMin string
	CreatedAtMax string
	UpdatedAtMin string
	UpdatedAtMax string
	IsCreatedAt  string
	IsUpdatedAt  string
}

type versionEventRow struct {
	ID            string `json:"id"`
	EventAt       string `json:"event_at,omitempty"`
	EventType     string `json:"event_type,omitempty"`
	EventItemType string `json:"event_item_type,omitempty"`
	EventItemID   string `json:"event_item_id,omitempty"`
	BrokerID      string `json:"broker_id,omitempty"`
}

func newVersionEventsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List version events",
		Long: `List version events with filtering and pagination.

Version events record change events that are exported to downstream integrations.

Output Columns:
  ID          Version event identifier
  EVENT AT    When the event occurred
  EVENT TYPE  Event type (create/update/destroy)
  ITEM TYPE   Event item resource type
  ITEM ID     Event item resource ID
  BROKER      Broker ID (if present)

Filters:
  --created-at-min  Filter by created-at on/after (ISO 8601)
  --created-at-max  Filter by created-at on/before (ISO 8601)
  --updated-at-min  Filter by updated-at on/after (ISO 8601)
  --updated-at-max  Filter by updated-at on/before (ISO 8601)
  --is-created-at   Filter by presence of created-at (true/false)
  --is-updated-at   Filter by presence of updated-at (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List version events
  xbe view version-events list

  # Filter by created-at range
  xbe view version-events list --created-at-min 2025-01-01T00:00:00Z --created-at-max 2025-01-31T23:59:59Z

  # Output as JSON
  xbe view version-events list --json`,
		Args: cobra.NoArgs,
		RunE: runVersionEventsList,
	}
	initVersionEventsListFlags(cmd)
	return cmd
}

func init() {
	versionEventsCmd.AddCommand(newVersionEventsListCmd())
}

func initVersionEventsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by presence of created-at (true/false)")
	cmd.Flags().String("is-updated-at", "", "Filter by presence of updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runVersionEventsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseVersionEventsListOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[version-events]", "event-at,event-type,broker,event-item")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is-created-at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[is-updated-at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/version-events", query)
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

	rows := buildVersionEventRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderVersionEventsTable(cmd, rows)
}

func parseVersionEventsListOptions(cmd *cobra.Command) (versionEventsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return versionEventsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
		IsCreatedAt:  isCreatedAt,
		IsUpdatedAt:  isUpdatedAt,
	}, nil
}

func buildVersionEventRows(resp jsonAPIResponse) []versionEventRow {
	rows := make([]versionEventRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := versionEventRow{
			ID:        resource.ID,
			EventAt:   formatDateTime(stringAttr(attrs, "event-at")),
			EventType: stringAttr(attrs, "event-type"),
		}

		row.BrokerID = relationshipIDFromMap(resource.Relationships, "broker")

		if rel, ok := resource.Relationships["event-item"]; ok && rel.Data != nil {
			row.EventItemID = rel.Data.ID
			row.EventItemType = rel.Data.Type
		}

		rows = append(rows, row)
	}
	return rows
}

func renderVersionEventsTable(cmd *cobra.Command, rows []versionEventRow) error {
	out := cmd.OutOrStdout()
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "ID\tEVENT AT\tEVENT TYPE\tITEM TYPE\tITEM ID\tBROKER")
	for _, row := range rows {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.EventAt,
			row.EventType,
			row.EventItemType,
			row.EventItemID,
			row.BrokerID,
		)
	}

	return w.Flush()
}
