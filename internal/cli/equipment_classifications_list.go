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

type equipmentClassificationsListOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	NoAuth             bool
	Limit              int
	Offset             int
	Name               string
	Abbreviation       string
	Brokers            string
	Parent             string
	MobilizationMethod string
	IsRoot             string
	WithChildren       string
	WithoutChildren    string
}

type equipmentClassificationRow struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Abbreviation       string `json:"abbreviation,omitempty"`
	MobilizationMethod string `json:"mobilization_method,omitempty"`
	ParentID           string `json:"parent_id,omitempty"`
	ParentName         string `json:"parent_name,omitempty"`
}

func newEquipmentClassificationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List equipment classifications",
		Long: `List equipment classifications with filtering and pagination.

Equipment classifications categorize equipment by type with their mobilization
requirements. They can be organized hierarchically.

Output Columns:
  ID             Equipment classification identifier
  NAME           Classification name (e.g., Excavator, Loader)
  ABBREVIATION   Short code
  MOBILIZATION   How equipment is transported (trailer, drive, etc.)
  PARENT         Parent classification name (if hierarchical)

Filters:
  --name                  Filter by name (partial match, case-insensitive)
  --abbreviation          Filter by abbreviation (partial match, case-insensitive)
  --brokers               Filter by broker ID
  --parent                Filter by parent classification ID
  --mobilization-method   Filter by mobilization method
  --is-root               Filter root classifications only (true/false)
  --with-children         Filter classifications with children (true/false)
  --without-children      Filter classifications without children (true/false)`,
		Example: `  # List all equipment classifications
  xbe view equipment-classifications list

  # Filter root classifications only
  xbe view equipment-classifications list --is-root true

  # Filter by mobilization method
  xbe view equipment-classifications list --mobilization-method trailer

  # Search by name
  xbe view equipment-classifications list --name "Excavator"

  # Output as JSON
  xbe view equipment-classifications list --json`,
		RunE: runEquipmentClassificationsList,
	}
	initEquipmentClassificationsListFlags(cmd)
	return cmd
}

func init() {
	equipmentClassificationsCmd.AddCommand(newEquipmentClassificationsListCmd())
}

func initEquipmentClassificationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("name", "", "Filter by name (partial match)")
	cmd.Flags().String("abbreviation", "", "Filter by abbreviation (partial match)")
	cmd.Flags().String("brokers", "", "Filter by broker ID")
	cmd.Flags().String("parent", "", "Filter by parent classification ID")
	cmd.Flags().String("mobilization-method", "", "Filter by mobilization method")
	cmd.Flags().String("is-root", "", "Filter root classifications only (true/false)")
	cmd.Flags().String("with-children", "", "Filter with children (true/false)")
	cmd.Flags().String("without-children", "", "Filter without children (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentClassificationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseEquipmentClassificationsListOptions(cmd)
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
	query.Set("fields[equipment-classifications]", "name,abbreviation,mobilization-method,parent")
	query.Set("include", "parent")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[name]", opts.Name)
	setFilterIfPresent(query, "filter[abbreviation]", opts.Abbreviation)
	setFilterIfPresent(query, "filter[brokers]", opts.Brokers)
	setFilterIfPresent(query, "filter[parent]", opts.Parent)
	setFilterIfPresent(query, "filter[mobilization-method]", opts.MobilizationMethod)
	setFilterIfPresent(query, "filter[is-root]", opts.IsRoot)
	setFilterIfPresent(query, "filter[with-children]", opts.WithChildren)
	setFilterIfPresent(query, "filter[without-children]", opts.WithoutChildren)

	body, _, err := client.Get(cmd.Context(), "/v1/equipment-classifications", query)
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

	rows := buildEquipmentClassificationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderEquipmentClassificationsTable(cmd, rows)
}

func parseEquipmentClassificationsListOptions(cmd *cobra.Command) (equipmentClassificationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	name, _ := cmd.Flags().GetString("name")
	abbreviation, _ := cmd.Flags().GetString("abbreviation")
	brokers, _ := cmd.Flags().GetString("brokers")
	parent, _ := cmd.Flags().GetString("parent")
	mobilizationMethod, _ := cmd.Flags().GetString("mobilization-method")
	isRoot, _ := cmd.Flags().GetString("is-root")
	withChildren, _ := cmd.Flags().GetString("with-children")
	withoutChildren, _ := cmd.Flags().GetString("without-children")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentClassificationsListOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		NoAuth:             noAuth,
		Limit:              limit,
		Offset:             offset,
		Name:               name,
		Abbreviation:       abbreviation,
		Brokers:            brokers,
		Parent:             parent,
		MobilizationMethod: mobilizationMethod,
		IsRoot:             isRoot,
		WithChildren:       withChildren,
		WithoutChildren:    withoutChildren,
	}, nil
}

func buildEquipmentClassificationRows(resp jsonAPIResponse) []equipmentClassificationRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]equipmentClassificationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := equipmentClassificationRow{
			ID:                 resource.ID,
			Name:               stringAttr(resource.Attributes, "name"),
			Abbreviation:       stringAttr(resource.Attributes, "abbreviation"),
			MobilizationMethod: stringAttr(resource.Attributes, "mobilization-method"),
		}

		if rel, ok := resource.Relationships["parent"]; ok && rel.Data != nil {
			row.ParentID = rel.Data.ID
			if parent, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.ParentName = stringAttr(parent.Attributes, "name")
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderEquipmentClassificationsTable(cmd *cobra.Command, rows []equipmentClassificationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No equipment classifications found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tABBREVIATION\tMOBILIZATION\tPARENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 30),
			truncateString(row.Abbreviation, 10),
			truncateString(row.MobilizationMethod, 15),
			truncateString(row.ParentName, 25),
		)
	}
	return writer.Flush()
}
