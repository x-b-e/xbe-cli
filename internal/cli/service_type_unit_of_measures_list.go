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

type serviceTypeUnitOfMeasuresListOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	NoAuth        bool
	Limit         int
	Offset        int
	Sort          string
	ServiceType   string
	UnitOfMeasure string
	Quantifiable  string
}

type serviceTypeUnitOfMeasureRow struct {
	ID                         string `json:"id"`
	Name                       string `json:"name,omitempty"`
	ServiceTypeID              string `json:"service_type_id,omitempty"`
	UnitOfMeasureID            string `json:"unit_of_measure_id,omitempty"`
	IsQuantifiable             bool   `json:"is_quantifiable"`
	IsSupplementalQuantifiable bool   `json:"is_supplemental_quantifiable"`
}

func newServiceTypeUnitOfMeasuresListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List service type unit of measures",
		Long: `List service type unit of measures.

Output Columns:
  ID            Service type unit of measure ID
  NAME          Service type unit of measure name
  SERVICE TYPE  Service type ID
  UNIT          Unit of measure ID
  QUANTIFIABLE  Whether this unit is quantifiable
  SUPP QUANT    Whether this unit is supplemental-quantifiable

Filters:
  --service-type     Filter by service type ID
  --unit-of-measure  Filter by unit of measure ID
  --quantifiable     Filter by quantifiable status (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List service type unit of measures
  xbe view service-type-unit-of-measures list

  # Filter by service type
  xbe view service-type-unit-of-measures list --service-type 123

  # Filter by unit of measure
  xbe view service-type-unit-of-measures list --unit-of-measure 456

  # Filter by quantifiable flag
  xbe view service-type-unit-of-measures list --quantifiable true

  # Output as JSON
  xbe view service-type-unit-of-measures list --json`,
		Args: cobra.NoArgs,
		RunE: runServiceTypeUnitOfMeasuresList,
	}
	initServiceTypeUnitOfMeasuresListFlags(cmd)
	return cmd
}

func init() {
	serviceTypeUnitOfMeasuresCmd.AddCommand(newServiceTypeUnitOfMeasuresListCmd())
}

func initServiceTypeUnitOfMeasuresListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("service-type", "", "Filter by service type ID")
	cmd.Flags().String("unit-of-measure", "", "Filter by unit of measure ID")
	cmd.Flags().String("quantifiable", "", "Filter by quantifiable status (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runServiceTypeUnitOfMeasuresList(cmd *cobra.Command, _ []string) error {
	opts, err := parseServiceTypeUnitOfMeasuresListOptions(cmd)
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
	query.Set("fields[service-type-unit-of-measures]", "name,is-quantifiable,is-supplemental-quantifiable,service-type,unit-of-measure")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[service-type]", opts.ServiceType)
	setFilterIfPresent(query, "filter[unit-of-measure]", opts.UnitOfMeasure)
	setFilterIfPresent(query, "filter[quantifiable]", opts.Quantifiable)

	body, _, err := client.Get(cmd.Context(), "/v1/service-type-unit-of-measures", query)
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

	rows := buildServiceTypeUnitOfMeasureRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderServiceTypeUnitOfMeasuresTable(cmd, rows)
}

func parseServiceTypeUnitOfMeasuresListOptions(cmd *cobra.Command) (serviceTypeUnitOfMeasuresListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	serviceType, _ := cmd.Flags().GetString("service-type")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	quantifiable, _ := cmd.Flags().GetString("quantifiable")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return serviceTypeUnitOfMeasuresListOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		NoAuth:        noAuth,
		Limit:         limit,
		Offset:        offset,
		Sort:          sort,
		ServiceType:   serviceType,
		UnitOfMeasure: unitOfMeasure,
		Quantifiable:  quantifiable,
	}, nil
}

func buildServiceTypeUnitOfMeasureRows(resp jsonAPIResponse) []serviceTypeUnitOfMeasureRow {
	rows := make([]serviceTypeUnitOfMeasureRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildServiceTypeUnitOfMeasureRow(resource)
		rows = append(rows, row)
	}
	return rows
}

func buildServiceTypeUnitOfMeasureRow(resource jsonAPIResource) serviceTypeUnitOfMeasureRow {
	row := serviceTypeUnitOfMeasureRow{
		ID:                         resource.ID,
		Name:                       stringAttr(resource.Attributes, "name"),
		IsQuantifiable:             boolAttr(resource.Attributes, "is-quantifiable"),
		IsSupplementalQuantifiable: boolAttr(resource.Attributes, "is-supplemental-quantifiable"),
	}

	if rel, ok := resource.Relationships["service-type"]; ok && rel.Data != nil {
		row.ServiceTypeID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["unit-of-measure"]; ok && rel.Data != nil {
		row.UnitOfMeasureID = rel.Data.ID
	}

	return row
}

func renderServiceTypeUnitOfMeasuresTable(cmd *cobra.Command, rows []serviceTypeUnitOfMeasureRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No service type unit of measures found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tSERVICE TYPE\tUNIT\tQUANTIFIABLE\tSUPP QUANT")
	for _, row := range rows {
		quantifiable := "no"
		if row.IsQuantifiable {
			quantifiable = "yes"
		}
		supplemental := "no"
		if row.IsSupplementalQuantifiable {
			supplemental = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 32),
			row.ServiceTypeID,
			row.UnitOfMeasureID,
			quantifiable,
			supplemental,
		)
	}
	return writer.Flush()
}
