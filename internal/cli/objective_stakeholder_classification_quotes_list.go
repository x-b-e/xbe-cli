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

type objectiveStakeholderClassificationQuotesListOptions struct {
	BaseURL                            string
	Token                              string
	JSON                               bool
	NoAuth                             bool
	Limit                              int
	Offset                             int
	Sort                               string
	ObjectiveStakeholderClassification string
	InterestDegreeMin                  string
	InterestDegreeMax                  string
}

type objectiveStakeholderClassificationQuoteRow struct {
	ID                                   string `json:"id"`
	Content                              string `json:"content,omitempty"`
	IsGenerated                          bool   `json:"is_generated"`
	ObjectiveStakeholderClassificationID string `json:"objective_stakeholder_classification_id,omitempty"`
}

func newObjectiveStakeholderClassificationQuotesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List objective stakeholder classification quotes",
		Long: `List objective stakeholder classification quotes.

Output Columns:
  ID              Quote identifier
  CONTENT         Quote content
  GENERATED       Whether the quote was generated
  CLASSIFICATION  Objective stakeholder classification ID

Filters:
  --objective-stakeholder-classification  Filter by objective stakeholder classification ID
  --interest-degree-min                   Filter by minimum interest degree
  --interest-degree-max                   Filter by maximum interest degree

Sort:
  --sort interest_degree  Sort by interest degree
  --sort random           Randomize ordering

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List quotes
  xbe view objective-stakeholder-classification-quotes list

  # Filter by classification
  xbe view objective-stakeholder-classification-quotes list --objective-stakeholder-classification 123

  # Filter by interest degree range
  xbe view objective-stakeholder-classification-quotes list --interest-degree-min 2 --interest-degree-max 5

  # Random order
  xbe view objective-stakeholder-classification-quotes list --sort random

  # JSON output
  xbe view objective-stakeholder-classification-quotes list --json`,
		Args: cobra.NoArgs,
		RunE: runObjectiveStakeholderClassificationQuotesList,
	}
	initObjectiveStakeholderClassificationQuotesListFlags(cmd)
	return cmd
}

func init() {
	objectiveStakeholderClassificationQuotesCmd.AddCommand(newObjectiveStakeholderClassificationQuotesListCmd())
}

func initObjectiveStakeholderClassificationQuotesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("objective-stakeholder-classification", "", "Filter by objective stakeholder classification ID")
	cmd.Flags().String("interest-degree-min", "", "Filter by minimum interest degree")
	cmd.Flags().String("interest-degree-max", "", "Filter by maximum interest degree")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runObjectiveStakeholderClassificationQuotesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseObjectiveStakeholderClassificationQuotesListOptions(cmd)
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
	query.Set("fields[objective-stakeholder-classification-quotes]", "content,is-generated,objective-stakeholder-classification")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[objective_stakeholder_classification]", opts.ObjectiveStakeholderClassification)
	setFilterIfPresent(query, "filter[interest_degree_min]", opts.InterestDegreeMin)
	setFilterIfPresent(query, "filter[interest_degree_max]", opts.InterestDegreeMax)

	body, _, err := client.Get(cmd.Context(), "/v1/objective-stakeholder-classification-quotes", query)
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

	rows := buildObjectiveStakeholderClassificationQuoteRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderObjectiveStakeholderClassificationQuotesTable(cmd, rows)
}

func parseObjectiveStakeholderClassificationQuotesListOptions(cmd *cobra.Command) (objectiveStakeholderClassificationQuotesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	objectiveStakeholderClassification, _ := cmd.Flags().GetString("objective-stakeholder-classification")
	interestDegreeMin, _ := cmd.Flags().GetString("interest-degree-min")
	interestDegreeMax, _ := cmd.Flags().GetString("interest-degree-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return objectiveStakeholderClassificationQuotesListOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		NoAuth:                             noAuth,
		Limit:                              limit,
		Offset:                             offset,
		Sort:                               sort,
		ObjectiveStakeholderClassification: objectiveStakeholderClassification,
		InterestDegreeMin:                  interestDegreeMin,
		InterestDegreeMax:                  interestDegreeMax,
	}, nil
}

func buildObjectiveStakeholderClassificationQuoteRows(resp jsonAPIResponse) []objectiveStakeholderClassificationQuoteRow {
	rows := make([]objectiveStakeholderClassificationQuoteRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildObjectiveStakeholderClassificationQuoteRow(resource)
		rows = append(rows, row)
	}
	return rows
}

func buildObjectiveStakeholderClassificationQuoteRow(resource jsonAPIResource) objectiveStakeholderClassificationQuoteRow {
	row := objectiveStakeholderClassificationQuoteRow{
		ID:          resource.ID,
		Content:     stringAttr(resource.Attributes, "content"),
		IsGenerated: boolAttr(resource.Attributes, "is-generated"),
	}

	if rel, ok := resource.Relationships["objective-stakeholder-classification"]; ok && rel.Data != nil {
		row.ObjectiveStakeholderClassificationID = rel.Data.ID
	}

	return row
}

func buildObjectiveStakeholderClassificationQuoteRowFromSingle(resp jsonAPISingleResponse) objectiveStakeholderClassificationQuoteRow {
	return buildObjectiveStakeholderClassificationQuoteRow(resp.Data)
}

func renderObjectiveStakeholderClassificationQuotesTable(cmd *cobra.Command, rows []objectiveStakeholderClassificationQuoteRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No objective stakeholder classification quotes found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tCONTENT\tGENERATED\tCLASSIFICATION")
	for _, row := range rows {
		generated := "no"
		if row.IsGenerated {
			generated = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			row.Content,
			generated,
			row.ObjectiveStakeholderClassificationID,
		)
	}
	return writer.Flush()
}
