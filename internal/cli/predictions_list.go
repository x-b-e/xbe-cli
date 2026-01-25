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

type predictionsListOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	NoAuth            bool
	Limit             int
	Offset            int
	Sort              string
	PredictionSubject string
	PredictedBy       string
	PredictionAgent   string
	Status            string
}

type predictionRow struct {
	ID                               string   `json:"id"`
	Status                           string   `json:"status,omitempty"`
	ContinuousRankedProbabilityScore *float64 `json:"continuous_ranked_probability_score,omitempty"`
	PredictionSubjectID              string   `json:"prediction_subject_id,omitempty"`
	PredictionSubjectName            string   `json:"prediction_subject_name,omitempty"`
	PredictionSubjectReferenceNumber string   `json:"prediction_subject_reference_number,omitempty"`
	PredictedByID                    string   `json:"predicted_by_id,omitempty"`
	PredictedByName                  string   `json:"predicted_by_name,omitempty"`
	PredictionAgentID                string   `json:"prediction_agent_id,omitempty"`
}

func newPredictionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List predictions",
		Long: `List predictions with filtering and pagination.

Predictions capture probability distributions for a prediction subject, along
with status and scoring metadata.

Output Columns:
  ID        Prediction identifier
  SUBJECT   Prediction subject name or ID
  PREDICTED Predicted-by user name or ID
  AGENT     Prediction agent ID
  STATUS    Prediction status
  CRPS      Continuous ranked probability score

Filters:
  --prediction-subject  Filter by prediction subject ID
  --predicted-by        Filter by predicted-by user ID
  --prediction-agent    Filter by prediction agent ID
  --status              Filter by status (draft, submitted, abandoned)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List predictions
  xbe view predictions list

  # Filter by prediction subject
  xbe view predictions list --prediction-subject 123

  # Filter by predicted-by user
  xbe view predictions list --predicted-by 456

  # Filter by status
  xbe view predictions list --status submitted

  # Output as JSON
  xbe view predictions list --json`,
		Args: cobra.NoArgs,
		RunE: runPredictionsList,
	}
	initPredictionsListFlags(cmd)
	return cmd
}

func init() {
	predictionsCmd.AddCommand(newPredictionsListCmd())
}

func initPredictionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("prediction-subject", "", "Filter by prediction subject ID")
	cmd.Flags().String("predicted-by", "", "Filter by predicted-by user ID")
	cmd.Flags().String("prediction-agent", "", "Filter by prediction agent ID")
	cmd.Flags().String("status", "", "Filter by status (draft, submitted, abandoned)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPredictionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parsePredictionsListOptions(cmd)
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
	query.Set("fields[predictions]", "status,continuous-ranked-probability-score,prediction-subject,predicted-by,prediction-agent")
	query.Set("include", "prediction-subject,predicted-by")
	query.Set("fields[prediction-subjects]", "name,reference-number")
	query.Set("fields[users]", "name,email-address")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[prediction-subject]", opts.PredictionSubject)
	setFilterIfPresent(query, "filter[predicted-by]", opts.PredictedBy)
	setFilterIfPresent(query, "filter[prediction-agent]", opts.PredictionAgent)
	setFilterIfPresent(query, "filter[status]", opts.Status)

	body, _, err := client.Get(cmd.Context(), "/v1/predictions", query)
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

	rows := buildPredictionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderPredictionsTable(cmd, rows)
}

func parsePredictionsListOptions(cmd *cobra.Command) (predictionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	predictionSubject, _ := cmd.Flags().GetString("prediction-subject")
	predictedBy, _ := cmd.Flags().GetString("predicted-by")
	predictionAgent, _ := cmd.Flags().GetString("prediction-agent")
	status, _ := cmd.Flags().GetString("status")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return predictionsListOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		NoAuth:            noAuth,
		Limit:             limit,
		Offset:            offset,
		Sort:              sort,
		PredictionSubject: predictionSubject,
		PredictedBy:       predictedBy,
		PredictionAgent:   predictionAgent,
		Status:            status,
	}, nil
}

func buildPredictionRows(resp jsonAPIResponse) []predictionRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]predictionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := predictionRow{ID: resource.ID}
		row.Status = stringAttr(resource.Attributes, "status")

		if score, ok := floatAttrValue(resource.Attributes, "continuous-ranked-probability-score"); ok {
			row.ContinuousRankedProbabilityScore = &score
		}

		if rel, ok := resource.Relationships["prediction-subject"]; ok && rel.Data != nil {
			row.PredictionSubjectID = rel.Data.ID
			if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.PredictionSubjectName = strings.TrimSpace(stringAttr(inc.Attributes, "name"))
				row.PredictionSubjectReferenceNumber = strings.TrimSpace(stringAttr(inc.Attributes, "reference-number"))
			}
		}

		if rel, ok := resource.Relationships["predicted-by"]; ok && rel.Data != nil {
			row.PredictedByID = rel.Data.ID
			if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				name := strings.TrimSpace(stringAttr(inc.Attributes, "name"))
				email := strings.TrimSpace(stringAttr(inc.Attributes, "email-address"))
				row.PredictedByName = firstNonEmpty(name, email)
			}
		}

		if rel, ok := resource.Relationships["prediction-agent"]; ok && rel.Data != nil {
			row.PredictionAgentID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderPredictionsTable(cmd *cobra.Command, rows []predictionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No predictions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSUBJECT\tPREDICTED BY\tAGENT\tSTATUS\tCRPS")
	for _, row := range rows {
		subjectLabel := firstNonEmpty(
			row.PredictionSubjectName,
			row.PredictionSubjectReferenceNumber,
			row.PredictionSubjectID,
		)
		predictedByLabel := firstNonEmpty(row.PredictedByName, row.PredictedByID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			subjectLabel,
			predictedByLabel,
			row.PredictionAgentID,
			row.Status,
			formatPredictionScore(row.ContinuousRankedProbabilityScore),
		)
	}
	return writer.Flush()
}

func predictionRowFromSingle(resp jsonAPISingleResponse) predictionRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	resource := resp.Data
	row := predictionRow{ID: resource.ID}
	row.Status = stringAttr(resource.Attributes, "status")

	if score, ok := floatAttrValue(resource.Attributes, "continuous-ranked-probability-score"); ok {
		row.ContinuousRankedProbabilityScore = &score
	}

	if rel, ok := resource.Relationships["prediction-subject"]; ok && rel.Data != nil {
		row.PredictionSubjectID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.PredictionSubjectName = strings.TrimSpace(stringAttr(inc.Attributes, "name"))
			row.PredictionSubjectReferenceNumber = strings.TrimSpace(stringAttr(inc.Attributes, "reference-number"))
		}
	}

	if rel, ok := resource.Relationships["predicted-by"]; ok && rel.Data != nil {
		row.PredictedByID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			name := strings.TrimSpace(stringAttr(inc.Attributes, "name"))
			email := strings.TrimSpace(stringAttr(inc.Attributes, "email-address"))
			row.PredictedByName = firstNonEmpty(name, email)
		}
	}

	if rel, ok := resource.Relationships["prediction-agent"]; ok && rel.Data != nil {
		row.PredictionAgentID = rel.Data.ID
	}

	return row
}

func formatPredictionScore(score *float64) string {
	if score == nil {
		return ""
	}
	return strconv.FormatFloat(*score, 'f', -1, 64)
}
