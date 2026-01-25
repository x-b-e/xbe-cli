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

type equipmentMovementStopCompletionsListOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	NoAuth         bool
	Limit          int
	Offset         int
	Sort           string
	Stop           string
	Equipment      string
	CompletedAtMin string
	CompletedAtMax string
	IsCompletedAt  string
	CreatedAtMin   string
	CreatedAtMax   string
	IsCreatedAt    string
	UpdatedAtMin   string
	UpdatedAtMax   string
	IsUpdatedAt    string
}

type equipmentMovementStopCompletionRow struct {
	ID          string `json:"id"`
	StopID      string `json:"stop_id,omitempty"`
	CompletedAt string `json:"completed_at,omitempty"`
	Latitude    string `json:"latitude,omitempty"`
	Longitude   string `json:"longitude,omitempty"`
	Note        string `json:"note,omitempty"`
	CreatedByID string `json:"created_by_id,omitempty"`
}

func newEquipmentMovementStopCompletionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List equipment movement stop completions",
		Long: `List equipment movement stop completions.

Output Columns:
  ID            Stop completion identifier
  STOP          Stop ID
  COMPLETED AT  Completion timestamp
  LAT           Latitude
  LNG           Longitude
  NOTE          Completion note
  CREATED BY    User who created the completion

Filters:
  --stop              Filter by stop ID
  --equipment         Filter by equipment ID
  --completed-at-min  Filter by completed-at on/after (ISO 8601)
  --completed-at-max  Filter by completed-at on/before (ISO 8601)
  --is-completed-at   Filter by presence of completed-at (true/false)
  --created-at-min    Filter by created-at on/after (ISO 8601)
  --created-at-max    Filter by created-at on/before (ISO 8601)
  --is-created-at     Filter by presence of created-at (true/false)
  --updated-at-min    Filter by updated-at on/after (ISO 8601)
  --updated-at-max    Filter by updated-at on/before (ISO 8601)
  --is-updated-at     Filter by presence of updated-at (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List stop completions
  xbe view equipment-movement-stop-completions list

  # Filter by stop
  xbe view equipment-movement-stop-completions list --stop 123

  # Filter by equipment
  xbe view equipment-movement-stop-completions list --equipment 456

  # Filter by completed-at range
  xbe view equipment-movement-stop-completions list \
    --completed-at-min 2026-01-22T00:00:00Z \
    --completed-at-max 2026-01-23T00:00:00Z

  # Output as JSON
  xbe view equipment-movement-stop-completions list --json`,
		Args: cobra.NoArgs,
		RunE: runEquipmentMovementStopCompletionsList,
	}
	initEquipmentMovementStopCompletionsListFlags(cmd)
	return cmd
}

func init() {
	equipmentMovementStopCompletionsCmd.AddCommand(newEquipmentMovementStopCompletionsListCmd())
}

func initEquipmentMovementStopCompletionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("stop", "", "Filter by stop ID")
	cmd.Flags().String("equipment", "", "Filter by equipment ID")
	cmd.Flags().String("completed-at-min", "", "Filter by completed-at on/after (ISO 8601)")
	cmd.Flags().String("completed-at-max", "", "Filter by completed-at on/before (ISO 8601)")
	cmd.Flags().String("is-completed-at", "", "Filter by presence of completed-at (true/false)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by presence of created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by presence of updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentMovementStopCompletionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseEquipmentMovementStopCompletionsListOptions(cmd)
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
	query.Set("fields[equipment-movement-stop-completions]", "completed-at,latitude,longitude,note,stop,created-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[stop]", opts.Stop)
	setFilterIfPresent(query, "filter[equipment]", opts.Equipment)
	setFilterIfPresent(query, "filter[completed-at-min]", opts.CompletedAtMin)
	setFilterIfPresent(query, "filter[completed-at-max]", opts.CompletedAtMax)
	setFilterIfPresent(query, "filter[is-completed-at]", opts.IsCompletedAt)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is-created-at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is-updated-at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/equipment-movement-stop-completions", query)
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

	rows := buildEquipmentMovementStopCompletionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderEquipmentMovementStopCompletionsTable(cmd, rows)
}

func parseEquipmentMovementStopCompletionsListOptions(cmd *cobra.Command) (equipmentMovementStopCompletionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	stop, _ := cmd.Flags().GetString("stop")
	equipment, _ := cmd.Flags().GetString("equipment")
	completedAtMin, _ := cmd.Flags().GetString("completed-at-min")
	completedAtMax, _ := cmd.Flags().GetString("completed-at-max")
	isCompletedAt, _ := cmd.Flags().GetString("is-completed-at")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentMovementStopCompletionsListOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		NoAuth:         noAuth,
		Limit:          limit,
		Offset:         offset,
		Sort:           sort,
		Stop:           stop,
		Equipment:      equipment,
		CompletedAtMin: completedAtMin,
		CompletedAtMax: completedAtMax,
		IsCompletedAt:  isCompletedAt,
		CreatedAtMin:   createdAtMin,
		CreatedAtMax:   createdAtMax,
		IsCreatedAt:    isCreatedAt,
		UpdatedAtMin:   updatedAtMin,
		UpdatedAtMax:   updatedAtMax,
		IsUpdatedAt:    isUpdatedAt,
	}, nil
}

func buildEquipmentMovementStopCompletionRows(resp jsonAPIResponse) []equipmentMovementStopCompletionRow {
	rows := make([]equipmentMovementStopCompletionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := equipmentMovementStopCompletionRow{
			ID:          resource.ID,
			CompletedAt: formatDateTime(stringAttr(resource.Attributes, "completed-at")),
			Latitude:    stringAttr(resource.Attributes, "latitude"),
			Longitude:   stringAttr(resource.Attributes, "longitude"),
			Note:        stringAttr(resource.Attributes, "note"),
		}

		if rel, ok := resource.Relationships["stop"]; ok && rel.Data != nil {
			row.StopID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
			row.CreatedByID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderEquipmentMovementStopCompletionsTable(cmd *cobra.Command, rows []equipmentMovementStopCompletionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No equipment movement stop completions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTOP\tCOMPLETED AT\tLAT\tLNG\tNOTE\tCREATED BY")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.StopID,
			truncateString(row.CompletedAt, 19),
			truncateString(row.Latitude, 12),
			truncateString(row.Longitude, 12),
			truncateString(row.Note, 32),
			row.CreatedByID,
		)
	}
	return writer.Flush()
}
