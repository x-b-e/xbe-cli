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

type profitImprovementCategoriesListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Name    string
}

func newProfitImprovementCategoriesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List profit improvement categories",
		Long: `List profit improvement categories with filtering and pagination.

Profit improvement categories organize profit improvements into groups for
reporting and analysis purposes.

Note: Profit improvement categories are read-only and cannot be created,
updated, or deleted through the API.

Output Columns:
  ID     Profit improvement category identifier
  NAME   Category name

Filters:
  --name  Filter by name (partial match, case-insensitive)`,
		Example: `  # List all profit improvement categories
  xbe view profit-improvement-categories list

  # Filter by name
  xbe view profit-improvement-categories list --name "safety"

  # Output as JSON
  xbe view profit-improvement-categories list --json`,
		RunE: runProfitImprovementCategoriesList,
	}
	initProfitImprovementCategoriesListFlags(cmd)
	return cmd
}

func init() {
	profitImprovementCategoriesCmd.AddCommand(newProfitImprovementCategoriesListCmd())
}

func initProfitImprovementCategoriesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("name", "", "Filter by name (partial match)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProfitImprovementCategoriesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProfitImprovementCategoriesListOptions(cmd)
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
	query.Set("fields[profit-improvement-categories]", "name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[name]", opts.Name)

	body, _, err := client.Get(cmd.Context(), "/v1/profit-improvement-categories", query)
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

	rows := buildProfitImprovementCategoryRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProfitImprovementCategoriesTable(cmd, rows)
}

func parseProfitImprovementCategoriesListOptions(cmd *cobra.Command) (profitImprovementCategoriesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	name, _ := cmd.Flags().GetString("name")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return profitImprovementCategoriesListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Name:    name,
	}, nil
}

type profitImprovementCategoryRow struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func buildProfitImprovementCategoryRows(resp jsonAPIResponse) []profitImprovementCategoryRow {
	rows := make([]profitImprovementCategoryRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := profitImprovementCategoryRow{
			ID:   resource.ID,
			Name: stringAttr(resource.Attributes, "name"),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderProfitImprovementCategoriesTable(cmd *cobra.Command, rows []profitImprovementCategoryRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No profit improvement categories found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\n",
			row.ID,
			truncateString(row.Name, 50),
		)
	}
	return writer.Flush()
}
