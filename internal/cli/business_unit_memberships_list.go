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

type businessUnitMembershipsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	BusinessUnit string
	Membership   string
	Kind         string
}

type businessUnitMembershipRow struct {
	ID               string `json:"id"`
	BusinessUnitID   string `json:"business_unit_id,omitempty"`
	BusinessUnitName string `json:"business_unit_name,omitempty"`
	MembershipID     string `json:"membership_id,omitempty"`
	Kind             string `json:"kind,omitempty"`
}

func newBusinessUnitMembershipsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List business unit memberships",
		Long: `List business unit memberships with filtering and pagination.

Business unit memberships associate broker memberships with specific business units.

Output Columns:
  ID             Business unit membership identifier
  BUSINESS UNIT  Business unit name (or ID if name unavailable)
  MEMBERSHIP     Membership ID
  KIND           Role (manager/technician/general)

Filters:
  --business-unit  Filter by business unit ID (comma-separated for multiple)
  --membership     Filter by membership ID (comma-separated for multiple)
  --kind           Filter by kind (manager/technician/general)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List business unit memberships
  xbe view business-unit-memberships list

  # Filter by business unit
  xbe view business-unit-memberships list --business-unit 123

  # Filter by membership
  xbe view business-unit-memberships list --membership 456

  # Filter by kind
  xbe view business-unit-memberships list --kind manager

  # Output as JSON
  xbe view business-unit-memberships list --json`,
		Args: cobra.NoArgs,
		RunE: runBusinessUnitMembershipsList,
	}
	initBusinessUnitMembershipsListFlags(cmd)
	return cmd
}

func init() {
	businessUnitMembershipsCmd.AddCommand(newBusinessUnitMembershipsListCmd())
}

func initBusinessUnitMembershipsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("business-unit", "", "Filter by business unit ID (comma-separated for multiple)")
	cmd.Flags().String("membership", "", "Filter by membership ID (comma-separated for multiple)")
	cmd.Flags().String("kind", "", "Filter by kind (manager/technician/general)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBusinessUnitMembershipsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseBusinessUnitMembershipsListOptions(cmd)
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
	query.Set("fields[business-unit-memberships]", "kind,business-unit,membership")
	query.Set("fields[business-units]", "company-name")
	query.Set("include", "business-unit")

	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[business_unit]", opts.BusinessUnit)
	setFilterIfPresent(query, "filter[membership]", opts.Membership)
	setFilterIfPresent(query, "filter[kind]", opts.Kind)

	body, _, err := client.Get(cmd.Context(), "/v1/business-unit-memberships", query)
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

	rows := buildBusinessUnitMembershipRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderBusinessUnitMembershipsTable(cmd, rows)
}

func parseBusinessUnitMembershipsListOptions(cmd *cobra.Command) (businessUnitMembershipsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	businessUnit, _ := cmd.Flags().GetString("business-unit")
	membership, _ := cmd.Flags().GetString("membership")
	kind, _ := cmd.Flags().GetString("kind")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return businessUnitMembershipsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		BusinessUnit: businessUnit,
		Membership:   membership,
		Kind:         kind,
	}, nil
}

func buildBusinessUnitMembershipRows(resp jsonAPIResponse) []businessUnitMembershipRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]businessUnitMembershipRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := businessUnitMembershipRow{
			ID:   resource.ID,
			Kind: stringAttr(attrs, "kind"),
		}

		if rel, ok := resource.Relationships["business-unit"]; ok && rel.Data != nil {
			row.BusinessUnitID = rel.Data.ID
			if businessUnit, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.BusinessUnitName = stringAttr(businessUnit.Attributes, "company-name")
			}
		}

		if rel, ok := resource.Relationships["membership"]; ok && rel.Data != nil {
			row.MembershipID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderBusinessUnitMembershipsTable(cmd *cobra.Command, rows []businessUnitMembershipRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No business unit memberships found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tBUSINESS_UNIT\tMEMBERSHIP\tKIND")
	for _, row := range rows {
		businessUnit := firstNonEmpty(row.BusinessUnitName, row.BusinessUnitID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(businessUnit, 30),
			truncateString(row.MembershipID, 20),
			row.Kind,
		)
	}
	return writer.Flush()
}
