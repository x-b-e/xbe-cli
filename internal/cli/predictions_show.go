package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type predictionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type predictionDetails struct {
	ID                               string   `json:"id"`
	Status                           string   `json:"status,omitempty"`
	ContinuousRankedProbabilityScore *float64 `json:"continuous_ranked_probability_score,omitempty"`
	ProbabilityDistribution          any      `json:"probability_distribution,omitempty"`
	PredictionSubjectID              string   `json:"prediction_subject_id,omitempty"`
	PredictionSubjectName            string   `json:"prediction_subject_name,omitempty"`
	PredictionSubjectReferenceNumber string   `json:"prediction_subject_reference_number,omitempty"`
	PredictedByID                    string   `json:"predicted_by_id,omitempty"`
	PredictedByName                  string   `json:"predicted_by_name,omitempty"`
	PredictionAgentID                string   `json:"prediction_agent_id,omitempty"`
}

func newPredictionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show prediction details",
		Long: `Show the full details of a prediction.

Output Fields:
  ID            Prediction identifier
  Status        Prediction status
  CRPS          Continuous ranked probability score
  Distribution  Probability distribution details
  Subject       Prediction subject name or ID
  Predicted By  Predicted-by user name or ID
  Agent         Prediction agent ID

Arguments:
  <id>    The prediction ID (required). Use the list command to find IDs.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a prediction
  xbe view predictions show 123

  # JSON output
  xbe view predictions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runPredictionsShow,
	}
	initPredictionsShowFlags(cmd)
	return cmd
}

func init() {
	predictionsCmd.AddCommand(newPredictionsShowCmd())
}

func initPredictionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPredictionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parsePredictionsShowOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("prediction id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[predictions]", "probability-distribution,status,continuous-ranked-probability-score,prediction-subject,predicted-by,prediction-agent")
	query.Set("include", "prediction-subject,predicted-by")
	query.Set("fields[prediction-subjects]", "name,reference-number")
	query.Set("fields[users]", "name,email-address")

	body, _, err := client.Get(cmd.Context(), "/v1/predictions/"+id, query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildPredictionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderPredictionDetails(cmd, details)
}

func parsePredictionsShowOptions(cmd *cobra.Command) (predictionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return predictionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildPredictionDetails(resp jsonAPISingleResponse) predictionDetails {
	attrs := resp.Data.Attributes
	details := predictionDetails{ID: resp.Data.ID}

	details.Status = stringAttr(attrs, "status")
	if score, ok := floatAttrValue(attrs, "continuous-ranked-probability-score"); ok {
		details.ContinuousRankedProbabilityScore = &score
	}
	if value, ok := attrs["probability-distribution"]; ok {
		details.ProbabilityDistribution = value
	}

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	if rel, ok := resp.Data.Relationships["prediction-subject"]; ok && rel.Data != nil {
		details.PredictionSubjectID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.PredictionSubjectName = strings.TrimSpace(stringAttr(inc.Attributes, "name"))
			details.PredictionSubjectReferenceNumber = strings.TrimSpace(stringAttr(inc.Attributes, "reference-number"))
		}
	}

	if rel, ok := resp.Data.Relationships["predicted-by"]; ok && rel.Data != nil {
		details.PredictedByID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			name := strings.TrimSpace(stringAttr(inc.Attributes, "name"))
			email := strings.TrimSpace(stringAttr(inc.Attributes, "email-address"))
			details.PredictedByName = firstNonEmpty(name, email)
		}
	}

	if rel, ok := resp.Data.Relationships["prediction-agent"]; ok && rel.Data != nil {
		details.PredictionAgentID = rel.Data.ID
	}

	return details
}

func renderPredictionDetails(cmd *cobra.Command, details predictionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.ContinuousRankedProbabilityScore != nil {
		fmt.Fprintf(out, "CRPS: %s\n", formatPredictionScore(details.ContinuousRankedProbabilityScore))
	}

	subjectLabel := firstNonEmpty(details.PredictionSubjectName, details.PredictionSubjectReferenceNumber)
	if subjectLabel != "" {
		fmt.Fprintf(out, "Prediction Subject: %s\n", subjectLabel)
	}
	if details.PredictionSubjectID != "" {
		fmt.Fprintf(out, "Prediction Subject ID: %s\n", details.PredictionSubjectID)
	}

	predictedByLabel := firstNonEmpty(details.PredictedByName)
	if predictedByLabel != "" {
		fmt.Fprintf(out, "Predicted By: %s\n", predictedByLabel)
	}
	if details.PredictedByID != "" {
		fmt.Fprintf(out, "Predicted By ID: %s\n", details.PredictedByID)
	}

	if details.PredictionAgentID != "" {
		fmt.Fprintf(out, "Prediction Agent ID: %s\n", details.PredictionAgentID)
	}

	fmt.Fprintln(out, "Probability Distribution:")
	formatted := formatProbabilityDistribution(details.ProbabilityDistribution)
	if formatted == "" {
		fmt.Fprintln(out, "  (none)")
	} else {
		fmt.Fprintln(out, indentPredictionLines(formatted, "  "))
	}

	return nil
}

func formatProbabilityDistribution(value any) string {
	if value == nil {
		return ""
	}
	pretty, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(pretty)
}

func indentPredictionLines(value, prefix string) string {
	if value == "" {
		return ""
	}
	return prefix + strings.ReplaceAll(value, "\n", "\n"+prefix)
}
