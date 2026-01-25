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

type hosAnnotationsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	Broker       string
	HosDay       string
	HosEvent     string
	CreatedAtMin string
	CreatedAtMax string
	UpdatedAtMin string
	UpdatedAtMax string
}

type hosAnnotationRow struct {
	ID           string `json:"id"`
	AnnotationAt string `json:"annotation_at,omitempty"`
	Comment      string `json:"comment,omitempty"`
	BrokerID     string `json:"broker_id,omitempty"`
	HosDayID     string `json:"hos_day_id,omitempty"`
	HosEventID   string `json:"hos_event_id,omitempty"`
}

const hosAnnotationsCommentMax = 40
const hosAnnotationsAnnotationAtMax = 19

func newHosAnnotationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List HOS annotations",
		Long: `List HOS annotations.

Output Columns:
  ID             Annotation identifier
  ANNOTATION AT  When the annotation was recorded
  COMMENT        Annotation comment (truncated)
  BROKER         Broker ID
  HOS DAY        HOS day ID
  HOS EVENT      HOS event ID

Filters:
  --broker          Filter by broker ID
  --hos-day         Filter by HOS day ID
  --hos-event       Filter by HOS event ID
  --created-at-min  Filter by created-at on/after (ISO 8601)
  --created-at-max  Filter by created-at on/before (ISO 8601)
  --updated-at-min  Filter by updated-at on/after (ISO 8601)
  --updated-at-max  Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List HOS annotations
  xbe view hos-annotations list

  # Filter by HOS day
  xbe view hos-annotations list --hos-day 123

  # Filter by HOS event
  xbe view hos-annotations list --hos-event 456

  # Filter by broker
  xbe view hos-annotations list --broker 789

  # Output as JSON
  xbe view hos-annotations list --json`,
		Args: cobra.NoArgs,
		RunE: runHosAnnotationsList,
	}
	initHosAnnotationsListFlags(cmd)
	return cmd
}

func init() {
	hosAnnotationsCmd.AddCommand(newHosAnnotationsListCmd())
}

func initHosAnnotationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("hos-day", "", "Filter by HOS day ID")
	cmd.Flags().String("hos-event", "", "Filter by HOS event ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runHosAnnotationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseHosAnnotationsListOptions(cmd)
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

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[hos-day]", opts.HosDay)
	setFilterIfPresent(query, "filter[hos-event]", opts.HosEvent)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/hos-annotations", query)
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

	rows := buildHosAnnotationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderHosAnnotationsTable(cmd, rows)
}

func parseHosAnnotationsListOptions(cmd *cobra.Command) (hosAnnotationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	hosDay, _ := cmd.Flags().GetString("hos-day")
	hosEvent, _ := cmd.Flags().GetString("hos-event")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return hosAnnotationsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		Broker:       broker,
		HosDay:       hosDay,
		HosEvent:     hosEvent,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
	}, nil
}

func buildHosAnnotationRows(resp jsonAPIResponse) []hosAnnotationRow {
	rows := make([]hosAnnotationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := hosAnnotationRow{
			ID:           resource.ID,
			AnnotationAt: formatDateTime(stringAttr(attrs, "annotation-at")),
			Comment:      strings.TrimSpace(stringAttr(attrs, "comment")),
		}

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["hos-day"]; ok && rel.Data != nil {
			row.HosDayID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["hos-event"]; ok && rel.Data != nil {
			row.HosEventID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func buildHosAnnotationRowFromSingle(resp jsonAPISingleResponse) hosAnnotationRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := hosAnnotationRow{
		ID:           resource.ID,
		AnnotationAt: formatDateTime(stringAttr(attrs, "annotation-at")),
		Comment:      strings.TrimSpace(stringAttr(attrs, "comment")),
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["hos-day"]; ok && rel.Data != nil {
		row.HosDayID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["hos-event"]; ok && rel.Data != nil {
		row.HosEventID = rel.Data.ID
	}

	return row
}

func renderHosAnnotationsTable(cmd *cobra.Command, rows []hosAnnotationRow) error {
	out := cmd.OutOrStdout()
	writer := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)

	fmt.Fprintln(writer, "ID\tANNOTATION AT\tCOMMENT\tBROKER\tHOS DAY\tHOS EVENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.AnnotationAt, hosAnnotationsAnnotationAtMax),
			truncateString(row.Comment, hosAnnotationsCommentMax),
			row.BrokerID,
			row.HosDayID,
			row.HosEventID,
		)
	}

	return writer.Flush()
}
