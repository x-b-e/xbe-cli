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

type incidentTagsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Slug    string
}

type incidentTagRow struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Kinds       string `json:"kinds,omitempty"`
}

func newIncidentTagsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List incident tags",
		Long: `List incident tags with filtering and pagination.

Incident tags are used to categorize and label safety incidents for
reporting and analysis purposes.

Output Columns:
  ID           Incident tag identifier
  SLUG         URL-friendly identifier
  NAME         Tag name
  DESCRIPTION  Tag description
  KINDS        Types of incidents this tag applies to

Filters:
  --slug  Filter by slug (exact match)`,
		Example: `  # List all incident tags
  xbe view incident-tags list

  # Filter by slug
  xbe view incident-tags list --slug "property-damage"

  # Output as JSON
  xbe view incident-tags list --json`,
		RunE: runIncidentTagsList,
	}
	initIncidentTagsListFlags(cmd)
	return cmd
}

func init() {
	incidentTagsCmd.AddCommand(newIncidentTagsListCmd())
}

func initIncidentTagsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("slug", "", "Filter by slug (exact match)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runIncidentTagsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseIncidentTagsListOptions(cmd)
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
	query.Set("fields[incident-tags]", "slug,name,description,kinds")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[slug]", opts.Slug)

	body, _, err := client.Get(cmd.Context(), "/v1/incident-tags", query)
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

	rows := buildIncidentTagRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderIncidentTagsTable(cmd, rows)
}

func parseIncidentTagsListOptions(cmd *cobra.Command) (incidentTagsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	slug, _ := cmd.Flags().GetString("slug")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return incidentTagsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Slug:    slug,
	}, nil
}

func buildIncidentTagRows(resp jsonAPIResponse) []incidentTagRow {
	rows := make([]incidentTagRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := incidentTagRow{
			ID:          resource.ID,
			Slug:        stringAttr(resource.Attributes, "slug"),
			Name:        stringAttr(resource.Attributes, "name"),
			Description: stringAttr(resource.Attributes, "description"),
			Kinds:       strings.Join(stringSliceAttr(resource.Attributes, "kinds"), ", "),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderIncidentTagsTable(cmd *cobra.Command, rows []incidentTagRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No incident tags found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSLUG\tNAME\tDESCRIPTION\tKINDS")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Slug, 20),
			truncateString(row.Name, 25),
			truncateString(row.Description, 30),
			truncateString(row.Kinds, 20),
		)
	}
	return writer.Flush()
}
