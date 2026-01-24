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

type doPredictionSubjectMembershipsUpdateOptions struct {
	BaseURL                                    string
	Token                                      string
	JSON                                       bool
	ID                                         string
	CanManageMemberships                       string
	CanSeePredictionsWithoutCreatingPrediction string
	CanUpdatePredictionConsensus               string
	CanUpdateOrDestroyOthersPredictions        string
	CanManageGaps                              string
	FlagsSet                                   map[string]bool
}

func newDoPredictionSubjectMembershipsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a prediction subject membership",
		Long: `Update an existing prediction subject membership.

Only the permission fields you specify will be updated. User and prediction
subject cannot be changed after creation.

Flags:
  --can-manage-memberships                    Can manage memberships (true/false)
  --can-see-predictions-without-creating-prediction
                                             Can see predictions without creating one (true/false)
  --can-update-prediction-consensus           Can update prediction consensus (true/false)
  --can-update-or-destroy-others-predictions  Can update or destroy others' predictions (true/false)
  --can-manage-gaps                           Can manage gaps (true/false)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update permission flags
  xbe do prediction-subject-memberships update 123 --can-manage-memberships true

  # Update multiple permissions
  xbe do prediction-subject-memberships update 123 \
    --can-update-prediction-consensus true \
    --can-manage-gaps true

  # Output as JSON
  xbe do prediction-subject-memberships update 123 --can-manage-memberships true --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoPredictionSubjectMembershipsUpdate,
	}
	initDoPredictionSubjectMembershipsUpdateFlags(cmd)
	return cmd
}

func init() {
	doPredictionSubjectMembershipsCmd.AddCommand(newDoPredictionSubjectMembershipsUpdateCmd())
}

func initDoPredictionSubjectMembershipsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("can-manage-memberships", "", "Can manage memberships (true/false)")
	cmd.Flags().String("can-see-predictions-without-creating-prediction", "", "Can see predictions without creating one (true/false)")
	cmd.Flags().String("can-update-prediction-consensus", "", "Can update prediction consensus (true/false)")
	cmd.Flags().String("can-update-or-destroy-others-predictions", "", "Can update or destroy others' predictions (true/false)")
	cmd.Flags().String("can-manage-gaps", "", "Can manage gaps (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPredictionSubjectMembershipsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoPredictionSubjectMembershipsUpdateOptions(cmd, args)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("prediction subject membership id is required")
	}

	if !hasAnyFlagSet(opts.FlagsSet) {
		err := fmt.Errorf("at least one permission flag is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.FlagsSet["can-manage-memberships"] {
		attributes["can-manage-memberships"] = opts.CanManageMemberships == "true"
	}
	if opts.FlagsSet["can-see-predictions-without-creating-prediction"] {
		attributes["can-see-predictions-without-creating-prediction"] = opts.CanSeePredictionsWithoutCreatingPrediction == "true"
	}
	if opts.FlagsSet["can-update-prediction-consensus"] {
		attributes["can-update-prediction-consensus"] = opts.CanUpdatePredictionConsensus == "true"
	}
	if opts.FlagsSet["can-update-or-destroy-others-predictions"] {
		attributes["can-update-or-destroy-others-predictions"] = opts.CanUpdateOrDestroyOthersPredictions == "true"
	}
	if opts.FlagsSet["can-manage-gaps"] {
		attributes["can-manage-gaps"] = opts.CanManageGaps == "true"
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "prediction-subject-memberships",
			"id":         id,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/prediction-subject-memberships/"+id, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated prediction subject membership %s\n", row.ID)
	return nil
}

func parseDoPredictionSubjectMembershipsUpdateOptions(cmd *cobra.Command, args []string) (doPredictionSubjectMembershipsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	canManageMemberships, _ := cmd.Flags().GetString("can-manage-memberships")
	canSeePredictionsWithoutCreatingPrediction, _ := cmd.Flags().GetString("can-see-predictions-without-creating-prediction")
	canUpdatePredictionConsensus, _ := cmd.Flags().GetString("can-update-prediction-consensus")
	canUpdateOrDestroyOthersPredictions, _ := cmd.Flags().GetString("can-update-or-destroy-others-predictions")
	canManageGaps, _ := cmd.Flags().GetString("can-manage-gaps")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	flagsSet := map[string]bool{
		"can-manage-memberships":                          cmd.Flags().Changed("can-manage-memberships"),
		"can-see-predictions-without-creating-prediction": cmd.Flags().Changed("can-see-predictions-without-creating-prediction"),
		"can-update-prediction-consensus":                 cmd.Flags().Changed("can-update-prediction-consensus"),
		"can-update-or-destroy-others-predictions":        cmd.Flags().Changed("can-update-or-destroy-others-predictions"),
		"can-manage-gaps":                                 cmd.Flags().Changed("can-manage-gaps"),
	}

	return doPredictionSubjectMembershipsUpdateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		ID:                   args[0],
		CanManageMemberships: canManageMemberships,
		CanSeePredictionsWithoutCreatingPrediction: canSeePredictionsWithoutCreatingPrediction,
		CanUpdatePredictionConsensus:               canUpdatePredictionConsensus,
		CanUpdateOrDestroyOthersPredictions:        canUpdateOrDestroyOthersPredictions,
		CanManageGaps:                              canManageGaps,
		FlagsSet:                                   flagsSet,
	}, nil
}
