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

type questionsListOptions struct {
	BaseURL                             string
	Token                               string
	JSON                                bool
	NoAuth                              bool
	Limit                               int
	Offset                              int
	Sort                                string
	Content                             string
	Source                              string
	Motivation                          string
	IgnoreOrganizationScopedNewsletters string
	IsTriaged                           string
	CreatedBy                           string
	AskedBy                             string
	AssignedTo                          string
	IsAssigned                          string
	WithoutFeedback                     string
	WithFeedback                        string
	WithoutRelatedContent               string
	WithRelatedContent                  string
}

type questionRow struct {
	ID           string `json:"id"`
	Content      string `json:"content,omitempty"`
	Source       string `json:"source,omitempty"`
	IsPublic     bool   `json:"is_public"`
	IsTriaged    bool   `json:"is_triaged"`
	AskedByID    string `json:"asked_by_id,omitempty"`
	CreatedByID  string `json:"created_by_id,omitempty"`
	AssignedToID string `json:"assigned_to_id,omitempty"`
}

func newQuestionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List questions",
		Long: `List questions with filtering and pagination.

Output Columns:
  ID           Question identifier
  CONTENT      Question content (truncated)
  SOURCE       Question source (app/link)
  PUBLIC       Whether the question is public
  TRIAGED      Whether the question has been triaged
  ASKED BY     User who asked the question
  CREATED BY   User who created the question
  ASSIGNED TO  User assigned to triage

Filters:
  --content                            Filter by question content
  --source                             Filter by source (app/link)
  --motivation                         Filter by motivation (serious/silly)
  --ignore-organization-scoped-newsletters  Filter by ignore newsletter scope (true/false)
  --is-triaged                          Filter by triage status (true/false)
  --created-by                          Filter by creator user ID
  --asked-by                            Filter by asking user ID
  --assigned-to                         Filter by assigned user ID
  --is-assigned                         Filter by assignment presence (true/false)
  --with-feedback                       Filter by feedback presence (true/false)
  --without-feedback                    Filter by lack of feedback (true/false)
  --with-related-content                Filter by related content presence (true/false)
  --without-related-content             Filter by lack of related content (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List questions
  xbe view questions list

  # Filter by source
  xbe view questions list --source app

  # Filter by assignee
  xbe view questions list --assigned-to 123

  # Output as JSON
  xbe view questions list --json`,
		Args: cobra.NoArgs,
		RunE: runQuestionsList,
	}
	initQuestionsListFlags(cmd)
	return cmd
}

func init() {
	questionsCmd.AddCommand(newQuestionsListCmd())
}

func initQuestionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("content", "", "Filter by question content")
	cmd.Flags().String("source", "", "Filter by source (app/link)")
	cmd.Flags().String("motivation", "", "Filter by motivation (serious/silly)")
	cmd.Flags().String("ignore-organization-scoped-newsletters", "", "Filter by ignore newsletter scope (true/false)")
	cmd.Flags().String("is-triaged", "", "Filter by triage status (true/false)")
	cmd.Flags().String("created-by", "", "Filter by creator user ID")
	cmd.Flags().String("asked-by", "", "Filter by asking user ID")
	cmd.Flags().String("assigned-to", "", "Filter by assigned user ID")
	cmd.Flags().String("is-assigned", "", "Filter by assignment presence (true/false)")
	cmd.Flags().String("with-feedback", "", "Filter by feedback presence (true/false)")
	cmd.Flags().String("without-feedback", "", "Filter by lack of feedback (true/false)")
	cmd.Flags().String("with-related-content", "", "Filter by related content presence (true/false)")
	cmd.Flags().String("without-related-content", "", "Filter by lack of related content (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runQuestionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseQuestionsListOptions(cmd)
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
	query.Set("fields[questions]", "content,source,is-public,is-triaged,asked-by,created-by,assigned-to")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[content]", opts.Content)
	setFilterIfPresent(query, "filter[source]", opts.Source)
	setFilterIfPresent(query, "filter[motivation]", opts.Motivation)
	setFilterIfPresent(query, "filter[ignore_organization_scoped_newsletters]", opts.IgnoreOrganizationScopedNewsletters)
	setFilterIfPresent(query, "filter[is_triaged]", opts.IsTriaged)
	setFilterIfPresent(query, "filter[created_by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[asked_by]", opts.AskedBy)
	setFilterIfPresent(query, "filter[assigned_to]", opts.AssignedTo)
	setFilterIfPresent(query, "filter[is_assigned]", opts.IsAssigned)
	setFilterIfPresent(query, "filter[with_feedback]", opts.WithFeedback)
	setFilterIfPresent(query, "filter[without_feedback]", opts.WithoutFeedback)
	setFilterIfPresent(query, "filter[with_related_content]", opts.WithRelatedContent)
	setFilterIfPresent(query, "filter[without_related_content]", opts.WithoutRelatedContent)

	body, _, err := client.Get(cmd.Context(), "/v1/questions", query)
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

	rows := buildQuestionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderQuestionsTable(cmd, rows)
}

func parseQuestionsListOptions(cmd *cobra.Command) (questionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	content, _ := cmd.Flags().GetString("content")
	source, _ := cmd.Flags().GetString("source")
	motivation, _ := cmd.Flags().GetString("motivation")
	ignoreOrganizationScopedNewsletters, _ := cmd.Flags().GetString("ignore-organization-scoped-newsletters")
	isTriaged, _ := cmd.Flags().GetString("is-triaged")
	createdBy, _ := cmd.Flags().GetString("created-by")
	askedBy, _ := cmd.Flags().GetString("asked-by")
	assignedTo, _ := cmd.Flags().GetString("assigned-to")
	isAssigned, _ := cmd.Flags().GetString("is-assigned")
	withoutFeedback, _ := cmd.Flags().GetString("without-feedback")
	withFeedback, _ := cmd.Flags().GetString("with-feedback")
	withoutRelatedContent, _ := cmd.Flags().GetString("without-related-content")
	withRelatedContent, _ := cmd.Flags().GetString("with-related-content")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return questionsListOptions{
		BaseURL:                             baseURL,
		Token:                               token,
		JSON:                                jsonOut,
		NoAuth:                              noAuth,
		Limit:                               limit,
		Offset:                              offset,
		Sort:                                sort,
		Content:                             content,
		Source:                              source,
		Motivation:                          motivation,
		IgnoreOrganizationScopedNewsletters: ignoreOrganizationScopedNewsletters,
		IsTriaged:                           isTriaged,
		CreatedBy:                           createdBy,
		AskedBy:                             askedBy,
		AssignedTo:                          assignedTo,
		IsAssigned:                          isAssigned,
		WithoutFeedback:                     withoutFeedback,
		WithFeedback:                        withFeedback,
		WithoutRelatedContent:               withoutRelatedContent,
		WithRelatedContent:                  withRelatedContent,
	}, nil
}

func buildQuestionRows(resp jsonAPIResponse) []questionRow {
	rows := make([]questionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildQuestionRow(resource))
	}
	return rows
}

func questionRowFromSingle(resp jsonAPISingleResponse) questionRow {
	return buildQuestionRow(resp.Data)
}

func buildQuestionRow(resource jsonAPIResource) questionRow {
	row := questionRow{
		ID:        resource.ID,
		Content:   stringAttr(resource.Attributes, "content"),
		Source:    stringAttr(resource.Attributes, "source"),
		IsPublic:  boolAttr(resource.Attributes, "is-public"),
		IsTriaged: boolAttr(resource.Attributes, "is-triaged"),
	}

	if rel, ok := resource.Relationships["asked-by"]; ok && rel.Data != nil {
		row.AskedByID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["assigned-to"]; ok && rel.Data != nil {
		row.AssignedToID = rel.Data.ID
	}

	return row
}

func renderQuestionsTable(cmd *cobra.Command, rows []questionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No questions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tCONTENT\tSOURCE\tPUBLIC\tTRIAGED\tASKED BY\tCREATED BY\tASSIGNED TO")
	for _, row := range rows {
		public := "no"
		if row.IsPublic {
			public = "yes"
		}
		triaged := "no"
		if row.IsTriaged {
			triaged = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Content, 60),
			row.Source,
			public,
			triaged,
			row.AskedByID,
			row.CreatedByID,
			row.AssignedToID,
		)
	}
	return writer.Flush()
}
