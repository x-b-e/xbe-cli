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

type doPredictionSubjectRecapGenerationsCreateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	PredictionSubjectID string
}

type predictionSubjectRecapGenerationRow struct {
	ID                  string `json:"id"`
	PredictionSubjectID string `json:"prediction_subject_id,omitempty"`
}

func newDoPredictionSubjectRecapGenerationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Generate a prediction subject recap",
		Long: `Generate a prediction subject recap.

This action schedules recap generation for the prediction subject.

Required flags:
  --prediction-subject   Prediction subject ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Generate a recap for a prediction subject
  xbe do prediction-subject-recap-generations create --prediction-subject 123

  # JSON output
  xbe do prediction-subject-recap-generations create --prediction-subject 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoPredictionSubjectRecapGenerationsCreate,
	}
	initDoPredictionSubjectRecapGenerationsCreateFlags(cmd)
	return cmd
}

func init() {
	doPredictionSubjectRecapGenerationsCmd.AddCommand(newDoPredictionSubjectRecapGenerationsCreateCmd())
}

func initDoPredictionSubjectRecapGenerationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("prediction-subject", "", "Prediction subject ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("prediction-subject")
}

func runDoPredictionSubjectRecapGenerationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoPredictionSubjectRecapGenerationsCreateOptions(cmd)
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

	relationships := map[string]any{
		"prediction-subject": map[string]any{
			"data": map[string]any{
				"type": "prediction-subjects",
				"id":   opts.PredictionSubjectID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "prediction-subject-recap-generations",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/prediction-subject-recap-generations", jsonBody)
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

	row := buildPredictionSubjectRecapGenerationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created prediction subject recap generation %s\n", row.ID)
	return nil
}

func parseDoPredictionSubjectRecapGenerationsCreateOptions(cmd *cobra.Command) (doPredictionSubjectRecapGenerationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	predictionSubjectID, _ := cmd.Flags().GetString("prediction-subject")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPredictionSubjectRecapGenerationsCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		PredictionSubjectID: predictionSubjectID,
	}, nil
}

func buildPredictionSubjectRecapGenerationRowFromSingle(resp jsonAPISingleResponse) predictionSubjectRecapGenerationRow {
	resource := resp.Data
	row := predictionSubjectRecapGenerationRow{
		ID: resource.ID,
	}
	if rel, ok := resource.Relationships["prediction-subject"]; ok && rel.Data != nil {
		row.PredictionSubjectID = rel.Data.ID
	}
	return row
}
