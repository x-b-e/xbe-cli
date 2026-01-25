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

type tagCategoriesListOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	NoAuth     bool
	Limit      int
	Offset     int
	Name       string
	Slug       string
	CanApplyTo string
}

func newTagCategoriesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tag categories",
		Long: `List tag categories with filtering and pagination.

Tag categories organize tags into groups based on what they can be applied to.

Output Columns:
  ID           Tag category identifier
  NAME         Category name
  SLUG         URL-friendly identifier
  CAN APPLY TO Entity types tags in this category can apply to

Filters:
  --name          Filter by name (partial match, case-insensitive)
  --slug          Filter by slug
  --can-apply-to  Filter by entity type (e.g., PredictionSubject, Comment)`,
		Example: `  # List all tag categories
  xbe view tag-categories list

  # Filter by name
  xbe view tag-categories list --name "market"

  # Filter by what they can apply to
  xbe view tag-categories list --can-apply-to PredictionSubject

  # Output as JSON
  xbe view tag-categories list --json`,
		RunE: runTagCategoriesList,
	}
	initTagCategoriesListFlags(cmd)
	return cmd
}

func init() {
	tagCategoriesCmd.AddCommand(newTagCategoriesListCmd())
}

func initTagCategoriesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("name", "", "Filter by name (partial match)")
	cmd.Flags().String("slug", "", "Filter by slug")
	cmd.Flags().String("can-apply-to", "", "Filter by entity type (e.g., PredictionSubject)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTagCategoriesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTagCategoriesListOptions(cmd)
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
	query.Set("fields[tag-categories]", "name,slug,description,can-apply-to")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[name]", opts.Name)
	setFilterIfPresent(query, "filter[slug]", opts.Slug)
	setFilterIfPresent(query, "filter[can-apply-to]", opts.CanApplyTo)

	body, _, err := client.Get(cmd.Context(), "/v1/tag-categories", query)
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

	rows := buildTagCategoryRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTagCategoriesTable(cmd, rows)
}

func parseTagCategoriesListOptions(cmd *cobra.Command) (tagCategoriesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	name, _ := cmd.Flags().GetString("name")
	slug, _ := cmd.Flags().GetString("slug")
	canApplyTo, _ := cmd.Flags().GetString("can-apply-to")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tagCategoriesListOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		NoAuth:     noAuth,
		Limit:      limit,
		Offset:     offset,
		Name:       name,
		Slug:       slug,
		CanApplyTo: canApplyTo,
	}, nil
}

func buildTagCategoryRows(resp jsonAPIResponse) []tagCategoryRow {
	rows := make([]tagCategoryRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := tagCategoryRow{
			ID:          resource.ID,
			Name:        stringAttr(resource.Attributes, "name"),
			Slug:        stringAttr(resource.Attributes, "slug"),
			Description: stringAttr(resource.Attributes, "description"),
			CanApplyTo:  stringSliceAttr(resource.Attributes, "can-apply-to"),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderTagCategoriesTable(cmd *cobra.Command, rows []tagCategoryRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No tag categories found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tSLUG\tCAN APPLY TO")
	for _, row := range rows {
		canApplyTo := strings.Join(row.CanApplyTo, ", ")
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 25),
			truncateString(row.Slug, 20),
			truncateString(canApplyTo, 40),
		)
	}
	return writer.Flush()
}
