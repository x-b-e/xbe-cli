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

type uiTourStepsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	NotID        string
	CreatedAtMin string
	CreatedAtMax string
	IsCreatedAt  string
	UpdatedAtMin string
	UpdatedAtMax string
	IsUpdatedAt  string
}

type uiTourStepRow struct {
	ID                 string `json:"id"`
	Name               string `json:"name,omitempty"`
	Abbreviation       string `json:"abbreviation,omitempty"`
	Sequence           string `json:"sequence,omitempty"`
	UiTourID           string `json:"ui_tour_id,omitempty"`
	UiTourName         string `json:"ui_tour_name,omitempty"`
	UiTourAbbreviation string `json:"ui_tour_abbreviation,omitempty"`
}

func newUiTourStepsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List UI tour steps",
		Long: `List UI tour steps with filtering and pagination.

Output Columns:
  ID            UI tour step identifier
  NAME          Step name
  ABBREVIATION  Step abbreviation
  SEQUENCE      Step sequence order
  UI TOUR       UI tour name or abbreviation

Filters:
  --not-id          Exclude UI tour step IDs (comma-separated)
  --created-at-min  Filter by created-at on/after (ISO 8601)
  --created-at-max  Filter by created-at on/before (ISO 8601)
  --is-created-at   Filter by has created-at (true/false)
  --updated-at-min  Filter by updated-at on/after (ISO 8601)
  --updated-at-max  Filter by updated-at on/before (ISO 8601)
  --is-updated-at   Filter by has updated-at (true/false)

Sorting:
  Default sort is by sequence. Use --sort to specify a different order.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List UI tour steps
  xbe view ui-tour-steps list

  # Sort by name
  xbe view ui-tour-steps list --sort name

  # Filter by created-at window
  xbe view ui-tour-steps list --created-at-min 2025-01-01T00:00:00Z --created-at-max 2025-12-31T23:59:59Z

  # Output as JSON
  xbe view ui-tour-steps list --json`,
		Args: cobra.NoArgs,
		RunE: runUiTourStepsList,
	}
	initUiTourStepsListFlags(cmd)
	return cmd
}

func init() {
	uiTourStepsCmd.AddCommand(newUiTourStepsListCmd())
}

func initUiTourStepsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("not-id", "", "Exclude UI tour step IDs (comma-separated)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by has created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by has updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runUiTourStepsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseUiTourStepsListOptions(cmd)
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
	query.Set("fields[ui-tour-steps]", "name,abbreviation,sequence,ui-tour")
	query.Set("include", "ui-tour")
	query.Set("fields[ui-tours]", "name,abbreviation")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	} else {
		query.Set("sort", "sequence")
	}

	setFilterIfPresent(query, "filter[not-id]", opts.NotID)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is-created-at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is-updated-at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/ui-tour-steps", query)
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

	rows := buildUiTourStepRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderUiTourStepsTable(cmd, rows)
}

func parseUiTourStepsListOptions(cmd *cobra.Command) (uiTourStepsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	notID, _ := cmd.Flags().GetString("not-id")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return uiTourStepsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		NotID:        notID,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		IsCreatedAt:  isCreatedAt,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
		IsUpdatedAt:  isUpdatedAt,
	}, nil
}

func buildUiTourStepRows(resp jsonAPIResponse) []uiTourStepRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]uiTourStepRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildUiTourStepRow(resource, included))
	}
	return rows
}

func uiTourStepRowFromSingle(resp jsonAPISingleResponse) uiTourStepRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildUiTourStepRow(resp.Data, included)
}

func buildUiTourStepRow(resource jsonAPIResource, included map[string]jsonAPIResource) uiTourStepRow {
	row := uiTourStepRow{
		ID:           resource.ID,
		Name:         strings.TrimSpace(stringAttr(resource.Attributes, "name")),
		Abbreviation: strings.TrimSpace(stringAttr(resource.Attributes, "abbreviation")),
		Sequence:     strings.TrimSpace(stringAttr(resource.Attributes, "sequence")),
	}

	if rel, ok := resource.Relationships["ui-tour"]; ok && rel.Data != nil {
		row.UiTourID = rel.Data.ID
		if tour, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.UiTourName = strings.TrimSpace(stringAttr(tour.Attributes, "name"))
			row.UiTourAbbreviation = strings.TrimSpace(stringAttr(tour.Attributes, "abbreviation"))
		}
	}

	return row
}

func renderUiTourStepsTable(cmd *cobra.Command, rows []uiTourStepRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No UI tour steps found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tABBREVIATION\tSEQUENCE\tUI TOUR")
	for _, row := range rows {
		uiTour := firstNonEmpty(row.UiTourName, row.UiTourAbbreviation, row.UiTourID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 30),
			truncateString(row.Abbreviation, 20),
			row.Sequence,
			truncateString(uiTour, 30),
		)
	}
	return writer.Flush()
}
