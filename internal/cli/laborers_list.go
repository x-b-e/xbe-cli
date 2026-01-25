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

type laborersListOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	NoAuth              bool
	Limit               int
	Offset              int
	LaborClassification string
	OrganizationType    string
	OrganizationID      string
	IsActive            string
	User                string
	CraftClass          string
	MissingCraftClass   string
	BusinessUnits       string
	Search              string
}

func newLaborersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List laborers",
		Long: `List laborers with filtering and pagination.

Laborers represent workers assigned to jobs and projects.

Output Columns:
  ID              Laborer identifier
  NICKNAME        Laborer nickname
  CLASSIFICATION  Labor classification ID
  USER ID         User ID
  ORG TYPE        Organization type
  ORG ID          Organization ID
  ACTIVE          Whether laborer is active

Filters:
  --labor-classification   Filter by labor classification ID
  --organization-type      Filter by organization type (e.g., Broker, Customer)
  --organization-id        Filter by organization ID
  --is-active              Filter by active status (true/false)
  --user                   Filter by user ID
  --craft-class            Filter by craft class ID
  --missing-craft-class    Filter by missing craft class (true/false)
  --business-units         Filter by business unit ID
  --search                 Search laborers`,
		Example: `  # List all laborers
  xbe view laborers list

  # Filter by labor classification
  xbe view laborers list --labor-classification 123

  # Filter by organization
  xbe view laborers list --organization-type Broker --organization-id 456

  # Filter by active status
  xbe view laborers list --is-active true

  # Search laborers
  xbe view laborers list --search "John"

  # Output as JSON
  xbe view laborers list --json`,
		RunE: runLaborersList,
	}
	initLaborersListFlags(cmd)
	return cmd
}

func init() {
	laborersCmd.AddCommand(newLaborersListCmd())
}

func initLaborersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("labor-classification", "", "Filter by labor classification ID")
	cmd.Flags().String("organization-type", "", "Filter by organization type (e.g., Broker, Customer)")
	cmd.Flags().String("organization-id", "", "Filter by organization ID")
	cmd.Flags().String("is-active", "", "Filter by active status (true/false)")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("craft-class", "", "Filter by craft class ID")
	cmd.Flags().String("missing-craft-class", "", "Filter by missing craft class (true/false)")
	cmd.Flags().String("business-units", "", "Filter by business unit ID")
	cmd.Flags().String("search", "", "Search laborers")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLaborersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseLaborersListOptions(cmd)
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
	query.Set("fields[laborers]", "nickname,is-active,labor-classification,user,organization")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[labor_classification]", opts.LaborClassification)
	if opts.OrganizationType != "" && opts.OrganizationID != "" {
		query.Set("filter[by_organization]", opts.OrganizationType+"|"+opts.OrganizationID)
	}
	setFilterIfPresent(query, "filter[is_active]", opts.IsActive)
	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[craft_class]", opts.CraftClass)
	setFilterIfPresent(query, "filter[missing_craft_class]", opts.MissingCraftClass)
	setFilterIfPresent(query, "filter[business_units]", opts.BusinessUnits)
	setFilterIfPresent(query, "filter[q]", opts.Search)

	body, _, err := client.Get(cmd.Context(), "/v1/laborers", query)
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

	rows := buildLaborerRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderLaborersTable(cmd, rows)
}

func parseLaborersListOptions(cmd *cobra.Command) (laborersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	laborClassification, _ := cmd.Flags().GetString("labor-classification")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	isActive, _ := cmd.Flags().GetString("is-active")
	user, _ := cmd.Flags().GetString("user")
	craftClass, _ := cmd.Flags().GetString("craft-class")
	missingCraftClass, _ := cmd.Flags().GetString("missing-craft-class")
	businessUnits, _ := cmd.Flags().GetString("business-units")
	search, _ := cmd.Flags().GetString("search")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return laborersListOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		NoAuth:              noAuth,
		Limit:               limit,
		Offset:              offset,
		LaborClassification: laborClassification,
		OrganizationType:    organizationType,
		OrganizationID:      organizationID,
		IsActive:            isActive,
		User:                user,
		CraftClass:          craftClass,
		MissingCraftClass:   missingCraftClass,
		BusinessUnits:       businessUnits,
		Search:              search,
	}, nil
}

type laborerRow struct {
	ID                    string `json:"id"`
	Nickname              string `json:"nickname,omitempty"`
	IsActive              bool   `json:"is_active,omitempty"`
	LaborClassificationID string `json:"labor_classification_id,omitempty"`
	UserID                string `json:"user_id,omitempty"`
	OrganizationType      string `json:"organization_type,omitempty"`
	OrganizationID        string `json:"organization_id,omitempty"`
}

func buildLaborerRows(resp jsonAPIResponse) []laborerRow {
	rows := make([]laborerRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := laborerRow{
			ID:       resource.ID,
			Nickname: stringAttr(resource.Attributes, "nickname"),
			IsActive: boolAttr(resource.Attributes, "is-active"),
		}

		if rel, ok := resource.Relationships["labor-classification"]; ok && rel.Data != nil {
			row.LaborClassificationID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
			row.UserID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
			row.OrganizationType = rel.Data.Type
			row.OrganizationID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderLaborersTable(cmd *cobra.Command, rows []laborerRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No laborers found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNICKNAME\tCLASSIFICATION\tUSER ID\tORG TYPE\tORG ID\tACTIVE")
	for _, row := range rows {
		active := "no"
		if row.IsActive {
			active = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Nickname, 25),
			row.LaborClassificationID,
			row.UserID,
			row.OrganizationType,
			row.OrganizationID,
			active,
		)
	}
	return writer.Flush()
}
