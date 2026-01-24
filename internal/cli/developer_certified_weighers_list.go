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

type developerCertifiedWeighersListOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	NoAuth    bool
	Limit     int
	Offset    int
	Sort      string
	Developer string
	User      string
	IsActive  string
}

type developerCertifiedWeigherRow struct {
	ID            string `json:"id"`
	Number        string `json:"number,omitempty"`
	IsActive      bool   `json:"is_active"`
	DeveloperID   string `json:"developer_id,omitempty"`
	DeveloperName string `json:"developer_name,omitempty"`
	UserID        string `json:"user_id,omitempty"`
	UserName      string `json:"user_name,omitempty"`
	UserEmail     string `json:"user_email,omitempty"`
}

func newDeveloperCertifiedWeighersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List developer certified weighers",
		Long: `List developer certified weighers with filtering and pagination.

Output Columns:
  ID          Certified weigher identifier
  NUMBER      Certified weigher number
  ACTIVE      Whether the weigher is active
  DEVELOPER   Developer name or ID
  USER        User name or email

Filters:
  --developer  Filter by developer ID
  --user       Filter by user ID
  --is-active  Filter by active status (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List developer certified weighers
  xbe view developer-certified-weighers list

  # Filter by developer
  xbe view developer-certified-weighers list --developer 123

  # Filter by user
  xbe view developer-certified-weighers list --user 456

  # Filter by active status
  xbe view developer-certified-weighers list --is-active true

  # Output as JSON
  xbe view developer-certified-weighers list --json`,
		Args: cobra.NoArgs,
		RunE: runDeveloperCertifiedWeighersList,
	}
	initDeveloperCertifiedWeighersListFlags(cmd)
	return cmd
}

func init() {
	developerCertifiedWeighersCmd.AddCommand(newDeveloperCertifiedWeighersListCmd())
}

func initDeveloperCertifiedWeighersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("developer", "", "Filter by developer ID")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("is-active", "", "Filter by active status (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDeveloperCertifiedWeighersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseDeveloperCertifiedWeighersListOptions(cmd)
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
	query.Set("fields[developer-certified-weighers]", "number,is-active,developer,user")
	query.Set("include", "developer,user")
	query.Set("fields[developers]", "name")
	query.Set("fields[users]", "name,email-address")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[developer]", opts.Developer)
	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[is_active]", opts.IsActive)

	body, _, err := client.Get(cmd.Context(), "/v1/developer-certified-weighers", query)
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

	rows := buildDeveloperCertifiedWeigherRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderDeveloperCertifiedWeighersTable(cmd, rows)
}

func parseDeveloperCertifiedWeighersListOptions(cmd *cobra.Command) (developerCertifiedWeighersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	developer, _ := cmd.Flags().GetString("developer")
	user, _ := cmd.Flags().GetString("user")
	isActive, _ := cmd.Flags().GetString("is-active")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return developerCertifiedWeighersListOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		NoAuth:    noAuth,
		Limit:     limit,
		Offset:    offset,
		Sort:      sort,
		Developer: developer,
		User:      user,
		IsActive:  isActive,
	}, nil
}

func buildDeveloperCertifiedWeigherRows(resp jsonAPIResponse) []developerCertifiedWeigherRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]developerCertifiedWeigherRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildDeveloperCertifiedWeigherRow(resource, included))
	}
	return rows
}

func developerCertifiedWeigherRowFromSingle(resp jsonAPISingleResponse) developerCertifiedWeigherRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildDeveloperCertifiedWeigherRow(resp.Data, included)
}

func buildDeveloperCertifiedWeigherRow(resource jsonAPIResource, included map[string]jsonAPIResource) developerCertifiedWeigherRow {
	row := developerCertifiedWeigherRow{
		ID:       resource.ID,
		Number:   stringAttr(resource.Attributes, "number"),
		IsActive: boolAttr(resource.Attributes, "is-active"),
	}

	if rel, ok := resource.Relationships["developer"]; ok && rel.Data != nil {
		row.DeveloperID = rel.Data.ID
		if developer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.DeveloperName = stringAttr(developer.Attributes, "name")
		}
	}

	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		row.UserID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.UserName = stringAttr(user.Attributes, "name")
			row.UserEmail = stringAttr(user.Attributes, "email-address")
		}
	}

	return row
}

func renderDeveloperCertifiedWeighersTable(cmd *cobra.Command, rows []developerCertifiedWeigherRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No developer certified weighers found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNUMBER\tACTIVE\tDEVELOPER\tUSER")
	for _, row := range rows {
		active := "no"
		if row.IsActive {
			active = "yes"
		}
		developerLabel := firstNonEmpty(row.DeveloperName, row.DeveloperID)
		userLabel := firstNonEmpty(row.UserName, row.UserEmail, row.UserID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Number, 20),
			active,
			truncateString(developerLabel, 30),
			truncateString(userLabel, 30),
		)
	}
	return writer.Flush()
}
