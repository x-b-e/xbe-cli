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

type doPredictionSubjectGapsCreateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	PredictionSubject string
	GapType           string
	Status            string
}

func newDoPredictionSubjectGapsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a prediction subject gap",
		Long: `Create a prediction subject gap.

Required:
  --prediction-subject  Prediction subject ID
  --gap-type            Gap type (actual_vs_walk_away, actual_vs_consensus, walk_away_vs_consensus)

Optional:
  --status              Status (pending, approved)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a prediction subject gap
  xbe do prediction-subject-gaps create \
    --prediction-subject 123 \
    --gap-type actual_vs_consensus

  # Create with explicit status
  xbe do prediction-subject-gaps create \
    --prediction-subject 123 \
    --gap-type actual_vs_walk_away \
    --status pending`,
		Args: cobra.NoArgs,
		RunE: runDoPredictionSubjectGapsCreate,
	}
	initDoPredictionSubjectGapsCreateFlags(cmd)
	return cmd
}

func init() {
	doPredictionSubjectGapsCmd.AddCommand(newDoPredictionSubjectGapsCreateCmd())
}

func initDoPredictionSubjectGapsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("prediction-subject", "", "Prediction subject ID (required)")
	cmd.Flags().String("gap-type", "", "Gap type (actual_vs_walk_away, actual_vs_consensus, walk_away_vs_consensus) (required)")
	cmd.Flags().String("status", "", "Status (pending, approved)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPredictionSubjectGapsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoPredictionSubjectGapsCreateOptions(cmd)
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
	if strings.TrimSpace(opts.GapType) == "" {
		err := fmt.Errorf("--gap-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"gap-type": opts.GapType,
	}
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "prediction-subject-gaps",
			"attributes": attributes,
			"relationships": map[string]any{
				"prediction-subject": map[string]any{
					"data": map[string]any{
						"type": "prediction-subjects",
						"id":   opts.PredictionSubject,
					},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/prediction-subject-gaps", jsonBody)
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

	row := predictionSubjectGapRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	message := fmt.Sprintf("Created prediction subject gap %s", row.ID)
	details := []string{}
	if row.GapType != "" {
		details = append(details, "type "+row.GapType)
	}
	if row.Status != "" {
		details = append(details, "status "+row.Status)
	}
	if row.GapAmount != nil {
		details = append(details, "gap "+formatPredictionSubjectGapAmount(row.GapAmount))
	}
	if len(details) > 0 {
		message = fmt.Sprintf("%s (%s)", message, strings.Join(details, ", "))
	}
	fmt.Fprintln(cmd.OutOrStdout(), message)
	return nil
}

func parseDoPredictionSubjectGapsCreateOptions(cmd *cobra.Command) (doPredictionSubjectGapsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	predictionSubject, _ := cmd.Flags().GetString("prediction-subject")
	gapType, _ := cmd.Flags().GetString("gap-type")
	status, _ := cmd.Flags().GetString("status")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPredictionSubjectGapsCreateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		PredictionSubject: predictionSubject,
		GapType:           gapType,
		Status:            status,
	}, nil
}
