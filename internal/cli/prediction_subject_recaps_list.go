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

type predictionSubjectRecapsListOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	NoAuth            bool
	Limit             int
	Offset            int
	Sort              string
	PredictionSubject string
}

type predictionSubjectRecapRow struct {
	ID                  string `json:"id"`
	PredictionSubjectID string `json:"prediction_subject_id,omitempty"`
	Markdown            string `json:"markdown,omitempty"`
}

func newPredictionSubjectRecapsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List prediction subject recaps",
		Long: `List prediction subject recaps with filtering and pagination.

Output Columns:
  ID       Prediction subject recap identifier
  SUBJECT  Prediction subject ID
  PREVIEW  Markdown preview (truncated)

Filters:
  --prediction-subject  Filter by prediction subject ID

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List prediction subject recaps
  xbe view prediction-subject-recaps list

  # Filter by prediction subject
  xbe view prediction-subject-recaps list --prediction-subject 123

  # Output as JSON
  xbe view prediction-subject-recaps list --json`,
		Args: cobra.NoArgs,
		RunE: runPredictionSubjectRecapsList,
	}
	initPredictionSubjectRecapsListFlags(cmd)
	return cmd
}

func init() {
	predictionSubjectRecapsCmd.AddCommand(newPredictionSubjectRecapsListCmd())
}

func initPredictionSubjectRecapsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("prediction-subject", "", "Filter by prediction subject ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPredictionSubjectRecapsList(cmd *cobra.Command, _ []string) error {
	opts, err := parsePredictionSubjectRecapsListOptions(cmd)
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
	query.Set("fields[prediction-subject-recaps]", "prediction-subject,markdown")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[prediction_subject]", opts.PredictionSubject)

	body, _, err := client.Get(cmd.Context(), "/v1/prediction-subject-recaps", query)
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

	rows := buildPredictionSubjectRecapRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderPredictionSubjectRecapsTable(cmd, rows)
}

func parsePredictionSubjectRecapsListOptions(cmd *cobra.Command) (predictionSubjectRecapsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	predictionSubject, _ := cmd.Flags().GetString("prediction-subject")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return predictionSubjectRecapsListOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		NoAuth:            noAuth,
		Limit:             limit,
		Offset:            offset,
		Sort:              sort,
		PredictionSubject: predictionSubject,
	}, nil
}

func buildPredictionSubjectRecapRows(resp jsonAPIResponse) []predictionSubjectRecapRow {
	rows := make([]predictionSubjectRecapRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildPredictionSubjectRecapRow(resource))
	}
	return rows
}

func predictionSubjectRecapRowFromSingle(resp jsonAPISingleResponse) predictionSubjectRecapRow {
	return buildPredictionSubjectRecapRow(resp.Data)
}

func buildPredictionSubjectRecapRow(resource jsonAPIResource) predictionSubjectRecapRow {
	row := predictionSubjectRecapRow{
		ID:       resource.ID,
		Markdown: stringAttr(resource.Attributes, "markdown"),
	}

	if rel, ok := resource.Relationships["prediction-subject"]; ok && rel.Data != nil {
		row.PredictionSubjectID = rel.Data.ID
	}

	return row
}

func renderPredictionSubjectRecapsTable(cmd *cobra.Command, rows []predictionSubjectRecapRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No prediction subject recaps found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSUBJECT\tPREVIEW")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			row.PredictionSubjectID,
			recapMarkdownPreview(row.Markdown),
		)
	}
	return writer.Flush()
}

func recapMarkdownPreview(markdown string) string {
	const previewMax = 60

	markdown = strings.TrimSpace(markdown)
	if markdown == "" {
		return ""
	}
	line := strings.SplitN(markdown, "\n", 2)[0]
	return truncateString(line, previewMax)
}
