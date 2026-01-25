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

type businessUnitLaborersListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	BusinessUnit string
	Laborer      string
}

type businessUnitLaborerRow struct {
	ID               string `json:"id"`
	BusinessUnitID   string `json:"business_unit_id,omitempty"`
	BusinessUnitName string `json:"business_unit_name,omitempty"`
	LaborerID        string `json:"laborer_id,omitempty"`
	LaborerName      string `json:"laborer_name,omitempty"`
}

func newBusinessUnitLaborersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List business unit laborers",
		Long: `List business unit laborers with filtering and pagination.

Output Columns:
  ID             Link identifier
  BUSINESS UNIT  Business unit name (falls back to ID)
  LABORER        Laborer nickname (falls back to ID)

Filters:
  --business-unit  Filter by business unit ID
  --laborer        Filter by laborer ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List business unit laborers
  xbe view business-unit-laborers list

  # Filter by business unit
  xbe view business-unit-laborers list --business-unit 123

  # Filter by laborer
  xbe view business-unit-laborers list --laborer 456

  # JSON output
  xbe view business-unit-laborers list --json`,
		Args: cobra.NoArgs,
		RunE: runBusinessUnitLaborersList,
	}
	initBusinessUnitLaborersListFlags(cmd)
	return cmd
}

func init() {
	businessUnitLaborersCmd.AddCommand(newBusinessUnitLaborersListCmd())
}

func initBusinessUnitLaborersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("business-unit", "", "Filter by business unit ID")
	cmd.Flags().String("laborer", "", "Filter by laborer ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBusinessUnitLaborersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseBusinessUnitLaborersListOptions(cmd)
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
	query.Set("fields[business-unit-laborers]", "business-unit,laborer")
	query.Set("fields[business-units]", "company-name")
	query.Set("fields[laborers]", "nickname")
	query.Set("include", "business-unit,laborer")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[business-unit]", opts.BusinessUnit)
	setFilterIfPresent(query, "filter[laborer]", opts.Laborer)

	body, _, err := client.Get(cmd.Context(), "/v1/business-unit-laborers", query)
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

	rows := buildBusinessUnitLaborerRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderBusinessUnitLaborersTable(cmd, rows)
}

func parseBusinessUnitLaborersListOptions(cmd *cobra.Command) (businessUnitLaborersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	businessUnit, _ := cmd.Flags().GetString("business-unit")
	laborer, _ := cmd.Flags().GetString("laborer")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return businessUnitLaborersListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		BusinessUnit: businessUnit,
		Laborer:      laborer,
	}, nil
}

func buildBusinessUnitLaborerRows(resp jsonAPIResponse) []businessUnitLaborerRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]businessUnitLaborerRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := businessUnitLaborerRow{
			ID: resource.ID,
		}

		if rel, ok := resource.Relationships["business-unit"]; ok && rel.Data != nil {
			row.BusinessUnitID = rel.Data.ID
			if unit, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.BusinessUnitName = strings.TrimSpace(stringAttr(unit.Attributes, "company-name"))
			}
		}

		if rel, ok := resource.Relationships["laborer"]; ok && rel.Data != nil {
			row.LaborerID = rel.Data.ID
			if laborer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.LaborerName = strings.TrimSpace(stringAttr(laborer.Attributes, "nickname"))
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderBusinessUnitLaborersTable(cmd *cobra.Command, rows []businessUnitLaborerRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No business unit laborers found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tBUSINESS UNIT\tLABORER")
	for _, row := range rows {
		businessUnitDisplay := firstNonEmpty(row.BusinessUnitName, row.BusinessUnitID)
		laborerDisplay := firstNonEmpty(row.LaborerName, row.LaborerID)
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			truncateString(businessUnitDisplay, 40),
			truncateString(laborerDisplay, 40),
		)
	}
	return writer.Flush()
}
