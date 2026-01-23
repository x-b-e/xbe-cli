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

type materialSiteUnavailabilitiesListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	MaterialSite string
	StartAtMin   string
	StartAtMax   string
	IsStartAt    string
	EndAtMin     string
	EndAtMax     string
	IsEndAt      string
	CreatedAtMin string
	CreatedAtMax string
	IsCreatedAt  string
	UpdatedAtMin string
	UpdatedAtMax string
	IsUpdatedAt  string
}

type materialSiteUnavailabilityRow struct {
	ID             string `json:"id"`
	MaterialSiteID string `json:"material_site_id,omitempty"`
	StartAt        string `json:"start_at,omitempty"`
	EndAt          string `json:"end_at,omitempty"`
	Description    string `json:"description,omitempty"`
}

func newMaterialSiteUnavailabilitiesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material site unavailabilities",
		Long: `List material site unavailabilities.

Output Columns:
  ID        Unavailability identifier
  SITE      Material site ID
  START AT  Start timestamp
  END AT    End timestamp
  DESC      Description

Filters:
  --material-site  Filter by material site ID
  --start-at-min   Filter by start-at on/after (ISO 8601)
  --start-at-max   Filter by start-at on/before (ISO 8601)
  --is-start-at    Filter by presence of start-at (true/false)
  --end-at-min     Filter by end-at on/after (ISO 8601)
  --end-at-max     Filter by end-at on/before (ISO 8601)
  --is-end-at      Filter by presence of end-at (true/false)
  --created-at-min Filter by created-at on/after (ISO 8601)
  --created-at-max Filter by created-at on/before (ISO 8601)
  --is-created-at  Filter by presence of created-at (true/false)
  --updated-at-min Filter by updated-at on/after (ISO 8601)
  --updated-at-max Filter by updated-at on/before (ISO 8601)
  --is-updated-at  Filter by presence of updated-at (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List material site unavailabilities
  xbe view material-site-unavailabilities list

  # Filter by material site
  xbe view material-site-unavailabilities list --material-site 123

  # Filter by start-at range
  xbe view material-site-unavailabilities list \
    --start-at-min 2026-01-23T00:00:00Z \
    --start-at-max 2026-01-24T00:00:00Z

  # Output as JSON
  xbe view material-site-unavailabilities list --json`,
		Args: cobra.NoArgs,
		RunE: runMaterialSiteUnavailabilitiesList,
	}
	initMaterialSiteUnavailabilitiesListFlags(cmd)
	return cmd
}

func init() {
	materialSiteUnavailabilitiesCmd.AddCommand(newMaterialSiteUnavailabilitiesListCmd())
}

func initMaterialSiteUnavailabilitiesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("material-site", "", "Filter by material site ID")
	cmd.Flags().String("start-at-min", "", "Filter by start-at on/after (ISO 8601)")
	cmd.Flags().String("start-at-max", "", "Filter by start-at on/before (ISO 8601)")
	cmd.Flags().String("is-start-at", "", "Filter by presence of start-at (true/false)")
	cmd.Flags().String("end-at-min", "", "Filter by end-at on/after (ISO 8601)")
	cmd.Flags().String("end-at-max", "", "Filter by end-at on/before (ISO 8601)")
	cmd.Flags().String("is-end-at", "", "Filter by presence of end-at (true/false)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by presence of created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by presence of updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialSiteUnavailabilitiesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialSiteUnavailabilitiesListOptions(cmd)
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
	query.Set("fields[material-site-unavailabilities]", "start-at,end-at,description,material-site")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[material_site]", opts.MaterialSite)
	setFilterIfPresent(query, "filter[start_at_min]", opts.StartAtMin)
	setFilterIfPresent(query, "filter[start_at_max]", opts.StartAtMax)
	setFilterIfPresent(query, "filter[is_start_at]", opts.IsStartAt)
	setFilterIfPresent(query, "filter[end_at_min]", opts.EndAtMin)
	setFilterIfPresent(query, "filter[end_at_max]", opts.EndAtMax)
	setFilterIfPresent(query, "filter[is_end_at]", opts.IsEndAt)
	setFilterIfPresent(query, "filter[created_at_min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created_at_max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is_created_at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated_at_min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated_at_max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is_updated_at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/material-site-unavailabilities", query)
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

	rows := buildMaterialSiteUnavailabilityRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialSiteUnavailabilitiesTable(cmd, rows)
}

func parseMaterialSiteUnavailabilitiesListOptions(cmd *cobra.Command) (materialSiteUnavailabilitiesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	materialSite, _ := cmd.Flags().GetString("material-site")
	startAtMin, _ := cmd.Flags().GetString("start-at-min")
	startAtMax, _ := cmd.Flags().GetString("start-at-max")
	isStartAt, _ := cmd.Flags().GetString("is-start-at")
	endAtMin, _ := cmd.Flags().GetString("end-at-min")
	endAtMax, _ := cmd.Flags().GetString("end-at-max")
	isEndAt, _ := cmd.Flags().GetString("is-end-at")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialSiteUnavailabilitiesListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		MaterialSite: materialSite,
		StartAtMin:   startAtMin,
		StartAtMax:   startAtMax,
		IsStartAt:    isStartAt,
		EndAtMin:     endAtMin,
		EndAtMax:     endAtMax,
		IsEndAt:      isEndAt,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		IsCreatedAt:  isCreatedAt,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
		IsUpdatedAt:  isUpdatedAt,
	}, nil
}

func buildMaterialSiteUnavailabilityRows(resp jsonAPIResponse) []materialSiteUnavailabilityRow {
	rows := make([]materialSiteUnavailabilityRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, materialSiteUnavailabilityRowFromResource(resource))
	}
	return rows
}

func materialSiteUnavailabilityRowFromResource(resource jsonAPIResource) materialSiteUnavailabilityRow {
	attrs := resource.Attributes
	row := materialSiteUnavailabilityRow{
		ID:          resource.ID,
		StartAt:     formatDateTime(stringAttr(attrs, "start-at")),
		EndAt:       formatDateTime(stringAttr(attrs, "end-at")),
		Description: stringAttr(attrs, "description"),
	}

	if rel, ok := resource.Relationships["material-site"]; ok && rel.Data != nil {
		row.MaterialSiteID = rel.Data.ID
	}

	return row
}

func buildMaterialSiteUnavailabilityRowFromSingle(resp jsonAPISingleResponse) materialSiteUnavailabilityRow {
	return materialSiteUnavailabilityRowFromResource(resp.Data)
}

func renderMaterialSiteUnavailabilitiesTable(cmd *cobra.Command, rows []materialSiteUnavailabilityRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material site unavailabilities found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSITE\tSTART AT\tEND AT\tDESC")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.MaterialSiteID,
			truncateString(row.StartAt, 19),
			truncateString(row.EndAt, 19),
			truncateString(row.Description, 24),
		)
	}
	return writer.Flush()
}
