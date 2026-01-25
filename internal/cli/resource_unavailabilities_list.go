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

type resourceUnavailabilitiesListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	ResourceType string
	ResourceID   string
	Organization string
	StartAtMin   string
	StartAtMax   string
	EndAtMin     string
	EndAtMax     string
}

type resourceUnavailabilityRow struct {
	ID           string `json:"id"`
	ResourceType string `json:"resource_type,omitempty"`
	ResourceID   string `json:"resource_id,omitempty"`
	StartAt      string `json:"start_at,omitempty"`
	EndAt        string `json:"end_at,omitempty"`
	Description  string `json:"description,omitempty"`
}

func newResourceUnavailabilitiesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List resource unavailabilities",
		Long: `List resource unavailabilities.

Output Columns:
  ID          Resource unavailability ID
  RESOURCE    Resource type and ID
  START AT    Start timestamp
  END AT      End timestamp
  DESCRIPTION Description

Filters:
  --resource-type   Filter by resource type (User, Equipment, Trailer, Tractor)
  --resource-id     Filter by resource ID (requires --resource-type)
  --organization    Filter by organization (Type|ID, e.g., Broker|123)
  --start-at-min    Filter by start-at on/after (ISO 8601)
  --start-at-max    Filter by start-at on/before (ISO 8601)
  --end-at-min      Filter by end-at on/after (ISO 8601)
  --end-at-max      Filter by end-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List resource unavailabilities
  xbe view resource-unavailabilities list

  # Filter by resource
  xbe view resource-unavailabilities list --resource-type User --resource-id 123

  # Filter by organization
  xbe view resource-unavailabilities list --organization "Broker|456"

  # Filter by time range
  xbe view resource-unavailabilities list --start-at-min 2025-01-01T00:00:00Z --end-at-max 2025-01-31T23:59:59Z

  # Output as JSON
  xbe view resource-unavailabilities list --json`,
		Args: cobra.NoArgs,
		RunE: runResourceUnavailabilitiesList,
	}
	initResourceUnavailabilitiesListFlags(cmd)
	return cmd
}

func init() {
	resourceUnavailabilitiesCmd.AddCommand(newResourceUnavailabilitiesListCmd())
}

func initResourceUnavailabilitiesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("resource-type", "", "Filter by resource type (User, Equipment, Trailer, Tractor)")
	cmd.Flags().String("resource-id", "", "Filter by resource ID (requires --resource-type)")
	cmd.Flags().String("organization", "", "Filter by organization (Type|ID, e.g., Broker|123)")
	cmd.Flags().String("start-at-min", "", "Filter by start-at on/after (ISO 8601)")
	cmd.Flags().String("start-at-max", "", "Filter by start-at on/before (ISO 8601)")
	cmd.Flags().String("end-at-min", "", "Filter by end-at on/after (ISO 8601)")
	cmd.Flags().String("end-at-max", "", "Filter by end-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runResourceUnavailabilitiesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseResourceUnavailabilitiesListOptions(cmd)
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

	if opts.ResourceID != "" && opts.ResourceType == "" {
		err := fmt.Errorf("--resource-type is required when --resource-id is set")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[resource-unavailabilities]", "start-at,end-at,description,resource,created-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	if opts.ResourceType != "" {
		resourceType := normalizeResourceTypeForFilter(opts.ResourceType)
		if opts.ResourceID != "" {
			query.Set("filter[resource]", resourceType+"|"+opts.ResourceID)
		} else {
			query.Set("filter[resource_type]", resourceType)
		}
	}
	setFilterIfPresent(query, "filter[organization]", opts.Organization)
	setFilterIfPresent(query, "filter[start_at_min]", opts.StartAtMin)
	setFilterIfPresent(query, "filter[start_at_max]", opts.StartAtMax)
	setFilterIfPresent(query, "filter[end_at_min]", opts.EndAtMin)
	setFilterIfPresent(query, "filter[end_at_max]", opts.EndAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/resource-unavailabilities", query)
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

	rows := buildResourceUnavailabilityRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderResourceUnavailabilitiesTable(cmd, rows)
}

func parseResourceUnavailabilitiesListOptions(cmd *cobra.Command) (resourceUnavailabilitiesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	resourceType, _ := cmd.Flags().GetString("resource-type")
	resourceID, _ := cmd.Flags().GetString("resource-id")
	organization, _ := cmd.Flags().GetString("organization")
	startAtMin, _ := cmd.Flags().GetString("start-at-min")
	startAtMax, _ := cmd.Flags().GetString("start-at-max")
	endAtMin, _ := cmd.Flags().GetString("end-at-min")
	endAtMax, _ := cmd.Flags().GetString("end-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return resourceUnavailabilitiesListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Organization: organization,
		StartAtMin:   startAtMin,
		StartAtMax:   startAtMax,
		EndAtMin:     endAtMin,
		EndAtMax:     endAtMax,
	}, nil
}

func buildResourceUnavailabilityRows(resp jsonAPIResponse) []resourceUnavailabilityRow {
	rows := make([]resourceUnavailabilityRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildResourceUnavailabilityRow(resource))
	}
	return rows
}

func buildResourceUnavailabilityRow(resource jsonAPIResource) resourceUnavailabilityRow {
	row := resourceUnavailabilityRow{
		ID:          resource.ID,
		StartAt:     formatDateTime(stringAttr(resource.Attributes, "start-at")),
		EndAt:       formatDateTime(stringAttr(resource.Attributes, "end-at")),
		Description: stringAttr(resource.Attributes, "description"),
	}

	if rel, ok := resource.Relationships["resource"]; ok && rel.Data != nil {
		row.ResourceType = rel.Data.Type
		row.ResourceID = rel.Data.ID
	}

	return row
}

func buildResourceUnavailabilityRowFromSingle(resp jsonAPISingleResponse) resourceUnavailabilityRow {
	return buildResourceUnavailabilityRow(resp.Data)
}

func renderResourceUnavailabilitiesTable(cmd *cobra.Command, rows []resourceUnavailabilityRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No resource unavailabilities found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tRESOURCE\tSTART AT\tEND AT\tDESCRIPTION")
	for _, row := range rows {
		resource := ""
		if row.ResourceType != "" && row.ResourceID != "" {
			resource = row.ResourceType + "/" + row.ResourceID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			resource,
			row.StartAt,
			row.EndAt,
			truncateString(row.Description, 40),
		)
	}
	return writer.Flush()
}
