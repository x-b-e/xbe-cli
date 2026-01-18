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

type glossaryTermsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Source  string
}

func newGlossaryTermsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List glossary terms",
		Long: `List glossary terms with filtering and pagination.

Returns a list of glossary terms matching the specified criteria.

Output Columns (table format):
  ID         Unique glossary term identifier
  TERM       The term being defined

Pagination:
  Use --limit and --offset to paginate through large result sets.

Filtering:
  Filter by source to see terms from specific origins.`,
		Example: `  # List all glossary terms
  xbe view glossary-terms list

  # Filter by source
  xbe view glossary-terms list --source xbe

  # Paginate results
  xbe view glossary-terms list --limit 20 --offset 40

  # Output as JSON for scripting
  xbe view glossary-terms list --json`,
		RunE: runGlossaryTermsList,
	}
	initGlossaryTermsListFlags(cmd)
	return cmd
}

func init() {
	glossaryTermsCmd.AddCommand(newGlossaryTermsListCmd())
}

func initGlossaryTermsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("source", "", "Filter by source")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runGlossaryTermsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseGlossaryTermsListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("sort", "term")
	query.Set("fields[glossary-terms]", "term,source")
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[source]", opts.Source)

	body, _, err := client.Get(cmd.Context(), "/v1/glossary-terms", query)
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

	if opts.JSON {
		rows := buildGlossaryTermRows(resp)
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderGlossaryTermsTable(cmd, resp)
}

func parseGlossaryTermsListOptions(cmd *cobra.Command) (glossaryTermsListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return glossaryTermsListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return glossaryTermsListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return glossaryTermsListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return glossaryTermsListOptions{}, err
	}
	source, err := cmd.Flags().GetString("source")
	if err != nil {
		return glossaryTermsListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return glossaryTermsListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return glossaryTermsListOptions{}, err
	}

	return glossaryTermsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Source:  source,
	}, nil
}

func renderGlossaryTermsTable(cmd *cobra.Command, resp jsonAPIResponse) error {
	rows := buildGlossaryTermRows(resp)
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No glossary terms found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, "ID\tTERM")

	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\n", row.ID, row.Term)
	}

	return writer.Flush()
}

type glossaryTermRow struct {
	ID     string `json:"id"`
	Term   string `json:"term"`
	Source string `json:"source"`
}

func buildGlossaryTermRows(resp jsonAPIResponse) []glossaryTermRow {
	rows := make([]glossaryTermRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, glossaryTermRow{
			ID:     resource.ID,
			Term:   strings.TrimSpace(stringAttr(resource.Attributes, "term")),
			Source: stringAttr(resource.Attributes, "source"),
		})
	}

	return rows
}
