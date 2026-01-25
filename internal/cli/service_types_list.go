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

type serviceTypesListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Name         string
	Abbreviation string
}

type serviceTypeRow struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Abbreviation     string `json:"abbreviation,omitempty"`
	IsManagementType bool   `json:"is_management_type"`
}

func newServiceTypesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List service types",
		Long: `List service types with filtering and pagination.

Service types define the types of services that can be performed on jobs
(e.g., hauling, spreading, compaction).

Output Columns:
  ID            Service type identifier
  NAME          Service type name
  ABBREVIATION  Short code
  MGMT TYPE     Whether this is a management type

Filters:
  --name          Filter by name (partial match, case-insensitive)
  --abbreviation  Filter by abbreviation (partial match, case-insensitive)`,
		Example: `  # List all service types
  xbe view service-types list

  # Search by name
  xbe view service-types list --name "haul"

  # Output as JSON
  xbe view service-types list --json`,
		RunE: runServiceTypesList,
	}
	initServiceTypesListFlags(cmd)
	return cmd
}

func init() {
	serviceTypesCmd.AddCommand(newServiceTypesListCmd())
}

func initServiceTypesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("name", "", "Filter by name (partial match)")
	cmd.Flags().String("abbreviation", "", "Filter by abbreviation (partial match)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runServiceTypesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseServiceTypesListOptions(cmd)
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
	query.Set("sort", "name")
	query.Set("fields[service-types]", "name,abbreviation,is-management-type")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[name]", opts.Name)
	setFilterIfPresent(query, "filter[abbreviation]", opts.Abbreviation)

	body, _, err := client.Get(cmd.Context(), "/v1/service-types", query)
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

	rows := buildServiceTypeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderServiceTypesTable(cmd, rows)
}

func parseServiceTypesListOptions(cmd *cobra.Command) (serviceTypesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	name, _ := cmd.Flags().GetString("name")
	abbreviation, _ := cmd.Flags().GetString("abbreviation")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return serviceTypesListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Name:         name,
		Abbreviation: abbreviation,
	}, nil
}

func buildServiceTypeRows(resp jsonAPIResponse) []serviceTypeRow {
	rows := make([]serviceTypeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := serviceTypeRow{
			ID:               resource.ID,
			Name:             stringAttr(resource.Attributes, "name"),
			Abbreviation:     stringAttr(resource.Attributes, "abbreviation"),
			IsManagementType: boolAttr(resource.Attributes, "is-management-type"),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderServiceTypesTable(cmd *cobra.Command, rows []serviceTypeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No service types found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tABBREVIATION\tMGMT TYPE")
	for _, row := range rows {
		mgmtType := "no"
		if row.IsManagementType {
			mgmtType = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 30),
			truncateString(row.Abbreviation, 10),
			mgmtType,
		)
	}
	return writer.Flush()
}
