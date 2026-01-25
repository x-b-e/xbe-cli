package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doPredictionsCreateOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	PredictionSubject       string
	PredictedBy             string
	ProbabilityDistribution string
	Status                  string
}

func newDoPredictionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a prediction",
		Long: `Create a prediction.

Required:
  --prediction-subject      Prediction subject ID

Optional:
  --predicted-by            Predicted-by user ID
  --status                  Status (draft, submitted, abandoned)
  --probability-distribution  JSON probability distribution payload (includes class_name)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a prediction
  xbe do predictions create \
    --prediction-subject 123 \
    --status draft \
    --probability-distribution '{"class_name":"TriangularDistribution","minimum":100,"mode":120,"maximum":140}'

  # Create a prediction for a specific user
  xbe do predictions create \
    --prediction-subject 123 \
    --predicted-by 456 \
    --status submitted`,
		Args: cobra.NoArgs,
		RunE: runDoPredictionsCreate,
	}
	initDoPredictionsCreateFlags(cmd)
	return cmd
}

func init() {
	doPredictionsCmd.AddCommand(newDoPredictionsCreateCmd())
}

func initDoPredictionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("prediction-subject", "", "Prediction subject ID (required)")
	cmd.Flags().String("predicted-by", "", "Predicted-by user ID")
	cmd.Flags().String("status", "", "Status (draft, submitted, abandoned)")
	cmd.Flags().String("probability-distribution", "", "JSON probability distribution payload")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPredictionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoPredictionsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
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

	if strings.TrimSpace(opts.PredictionSubject) == "" {
		err := fmt.Errorf("--prediction-subject is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("probability-distribution") {
		if opts.ProbabilityDistribution != "" {
			var distribution any
			if err := json.Unmarshal([]byte(opts.ProbabilityDistribution), &distribution); err != nil {
				err = fmt.Errorf("invalid probability-distribution JSON: %w", err)
				fmt.Fprintln(cmd.ErrOrStderr(), err)
				return err
			}
			attributes["probability-distribution"] = distribution
		} else {
			attributes["probability-distribution"] = nil
		}
	}

	relationships := map[string]any{
		"prediction-subject": map[string]any{
			"data": map[string]any{
				"type": "prediction-subjects",
				"id":   opts.PredictionSubject,
			},
		},
	}

	if strings.TrimSpace(opts.PredictedBy) != "" {
		relationships["predicted-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.PredictedBy,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "predictions",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/predictions", jsonBody)
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

	row := predictionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	message := fmt.Sprintf("Created prediction %s", row.ID)
	parts := []string{}
	if row.Status != "" {
		parts = append(parts, "status "+row.Status)
	}
	if row.PredictionSubjectID != "" {
		parts = append(parts, "subject "+row.PredictionSubjectID)
	}
	if len(parts) > 0 {
		message = fmt.Sprintf("%s (%s)", message, strings.Join(parts, ", "))
	}
	fmt.Fprintln(cmd.OutOrStdout(), message)
	return nil
}

func parseDoPredictionsCreateOptions(cmd *cobra.Command) (doPredictionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	predictionSubject, _ := cmd.Flags().GetString("prediction-subject")
	predictedBy, _ := cmd.Flags().GetString("predicted-by")
	status, _ := cmd.Flags().GetString("status")
	probabilityDistribution, _ := cmd.Flags().GetString("probability-distribution")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPredictionsCreateOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		PredictionSubject:       predictionSubject,
		PredictedBy:             predictedBy,
		Status:                  status,
		ProbabilityDistribution: probabilityDistribution,
	}, nil
}
