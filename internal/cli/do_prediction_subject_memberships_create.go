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

type doPredictionSubjectMembershipsCreateOptions struct {
	BaseURL                                    string
	Token                                      string
	JSON                                       bool
	PredictionSubject                          string
	User                                       string
	CreatedBy                                  string
	CanManageMemberships                       string
	CanSeePredictionsWithoutCreatingPrediction string
	CanUpdatePredictionConsensus               string
	CanUpdateOrDestroyOthersPredictions        string
	CanManageGaps                              string
}

func newDoPredictionSubjectMembershipsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a prediction subject membership",
		Long: `Create a prediction subject membership.

Required flags:
  --prediction-subject  Prediction subject ID (required)
  --user                User ID (required)

Optional flags:
  --created-by                                Creator user ID
  --can-manage-memberships                    Can manage memberships (true/false)
  --can-see-predictions-without-creating-prediction
                                               Can see predictions without creating one (true/false)
  --can-update-prediction-consensus           Can update prediction consensus (true/false)
  --can-update-or-destroy-others-predictions  Can update or destroy others' predictions (true/false)
  --can-manage-gaps                           Can manage gaps (true/false)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a prediction subject membership
  xbe do prediction-subject-memberships create --prediction-subject 123 --user 456

  # Create with permissions
  xbe do prediction-subject-memberships create \
    --prediction-subject 123 \
    --user 456 \
    --can-manage-memberships true \
    --can-update-prediction-consensus true

  # Output as JSON
  xbe do prediction-subject-memberships create --prediction-subject 123 --user 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoPredictionSubjectMembershipsCreate,
	}
	initDoPredictionSubjectMembershipsCreateFlags(cmd)
	return cmd
}

func init() {
	doPredictionSubjectMembershipsCmd.AddCommand(newDoPredictionSubjectMembershipsCreateCmd())
}

func initDoPredictionSubjectMembershipsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("prediction-subject", "", "Prediction subject ID (required)")
	cmd.Flags().String("user", "", "User ID (required)")
	cmd.Flags().String("created-by", "", "Creator user ID")
	cmd.Flags().String("can-manage-memberships", "", "Can manage memberships (true/false)")
	cmd.Flags().String("can-see-predictions-without-creating-prediction", "", "Can see predictions without creating one (true/false)")
	cmd.Flags().String("can-update-prediction-consensus", "", "Can update prediction consensus (true/false)")
	cmd.Flags().String("can-update-or-destroy-others-predictions", "", "Can update or destroy others' predictions (true/false)")
	cmd.Flags().String("can-manage-gaps", "", "Can manage gaps (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPredictionSubjectMembershipsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoPredictionSubjectMembershipsCreateOptions(cmd)
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

	if opts.PredictionSubject == "" {
		err := fmt.Errorf("--prediction-subject is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.User == "" {
		err := fmt.Errorf("--user is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	setBoolAttrIfPresent(attributes, "can-manage-memberships", opts.CanManageMemberships)
	setBoolAttrIfPresent(attributes, "can-see-predictions-without-creating-prediction", opts.CanSeePredictionsWithoutCreatingPrediction)
	setBoolAttrIfPresent(attributes, "can-update-prediction-consensus", opts.CanUpdatePredictionConsensus)
	setBoolAttrIfPresent(attributes, "can-update-or-destroy-others-predictions", opts.CanUpdateOrDestroyOthersPredictions)
	setBoolAttrIfPresent(attributes, "can-manage-gaps", opts.CanManageGaps)

	relationships := map[string]any{
		"prediction-subject": map[string]any{
			"data": map[string]any{
				"type": "prediction-subjects",
				"id":   opts.PredictionSubject,
			},
		},
		"user": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.User,
			},
		},
	}

	if opts.CreatedBy != "" {
		relationships["created-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.CreatedBy,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "prediction-subject-memberships",
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

	body, _, err := client.Post(cmd.Context(), "/v1/prediction-subject-memberships", jsonBody)
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

	row := predictionSubjectMembershipRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created prediction subject membership %s\n", row.ID)
	return nil
}

func parseDoPredictionSubjectMembershipsCreateOptions(cmd *cobra.Command) (doPredictionSubjectMembershipsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	predictionSubject, _ := cmd.Flags().GetString("prediction-subject")
	user, _ := cmd.Flags().GetString("user")
	createdBy, _ := cmd.Flags().GetString("created-by")
	canManageMemberships, _ := cmd.Flags().GetString("can-manage-memberships")
	canSeePredictionsWithoutCreatingPrediction, _ := cmd.Flags().GetString("can-see-predictions-without-creating-prediction")
	canUpdatePredictionConsensus, _ := cmd.Flags().GetString("can-update-prediction-consensus")
	canUpdateOrDestroyOthersPredictions, _ := cmd.Flags().GetString("can-update-or-destroy-others-predictions")
	canManageGaps, _ := cmd.Flags().GetString("can-manage-gaps")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPredictionSubjectMembershipsCreateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		PredictionSubject:    predictionSubject,
		User:                 user,
		CreatedBy:            createdBy,
		CanManageMemberships: canManageMemberships,
		CanSeePredictionsWithoutCreatingPrediction: canSeePredictionsWithoutCreatingPrediction,
		CanUpdatePredictionConsensus:               canUpdatePredictionConsensus,
		CanUpdateOrDestroyOthersPredictions:        canUpdateOrDestroyOthersPredictions,
		CanManageGaps:                              canManageGaps,
	}, nil
}
