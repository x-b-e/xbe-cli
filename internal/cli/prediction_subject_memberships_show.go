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

type predictionSubjectMembershipsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type predictionSubjectMembershipDetails struct {
	ID                                         string `json:"id"`
	PredictionSubjectID                        string `json:"prediction_subject_id,omitempty"`
	UserID                                     string `json:"user_id,omitempty"`
	CreatedByID                                string `json:"created_by_id,omitempty"`
	CanManageMemberships                       bool   `json:"can_manage_memberships"`
	CanSeePredictionsWithoutCreatingPrediction bool   `json:"can_see_predictions_without_creating_prediction"`
	CanUpdatePredictionConsensus               bool   `json:"can_update_prediction_consensus"`
	CanUpdateOrDestroyOthersPredictions        bool   `json:"can_update_or_destroy_others_predictions"`
	CanManageGaps                              bool   `json:"can_manage_gaps"`
}

func newPredictionSubjectMembershipsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show prediction subject membership details",
		Long: `Show the full details of a prediction subject membership.

Output Fields:
  ID             Prediction subject membership identifier
  Prediction Subject
  User
  Created By
  Can Manage Memberships
  Can See Predictions Without Creating Prediction
  Can Update Prediction Consensus
  Can Update or Destroy Others' Predictions
  Can Manage Gaps

Arguments:
  <id>  The prediction subject membership ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show prediction subject membership details
  xbe view prediction-subject-memberships show 123

  # Output as JSON
  xbe view prediction-subject-memberships show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runPredictionSubjectMembershipsShow,
	}
	initPredictionSubjectMembershipsShowFlags(cmd)
	return cmd
}

func init() {
	predictionSubjectMembershipsCmd.AddCommand(newPredictionSubjectMembershipsShowCmd())
}

func initPredictionSubjectMembershipsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPredictionSubjectMembershipsShow(cmd *cobra.Command, args []string) error {
	opts, err := parsePredictionSubjectMembershipsShowOptions(cmd)
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
		return fmt.Errorf("prediction subject membership id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[prediction-subject-memberships]", "can-manage-memberships,can-see-predictions-without-creating-prediction,can-update-prediction-consensus,can-update-or-destroy-others-predictions,can-manage-gaps,prediction-subject,user,created-by")

	body, _, err := client.Get(cmd.Context(), "/v1/prediction-subject-memberships/"+id, query)
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

	details := buildPredictionSubjectMembershipDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderPredictionSubjectMembershipDetails(cmd, details)
}

func parsePredictionSubjectMembershipsShowOptions(cmd *cobra.Command) (predictionSubjectMembershipsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return predictionSubjectMembershipsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildPredictionSubjectMembershipDetails(resp jsonAPISingleResponse) predictionSubjectMembershipDetails {
	attrs := resp.Data.Attributes
	data := resp.Data

	details := predictionSubjectMembershipDetails{
		ID:                   data.ID,
		CanManageMemberships: boolAttr(attrs, "can-manage-memberships"),
		CanSeePredictionsWithoutCreatingPrediction: boolAttr(attrs, "can-see-predictions-without-creating-prediction"),
		CanUpdatePredictionConsensus:               boolAttr(attrs, "can-update-prediction-consensus"),
		CanUpdateOrDestroyOthersPredictions:        boolAttr(attrs, "can-update-or-destroy-others-predictions"),
		CanManageGaps:                              boolAttr(attrs, "can-manage-gaps"),
	}

	if rel, ok := data.Relationships["prediction-subject"]; ok && rel.Data != nil {
		details.PredictionSubjectID = rel.Data.ID
	}
	if rel, ok := data.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
	}
	if rel, ok := data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}

	return details
}

func renderPredictionSubjectMembershipDetails(cmd *cobra.Command, details predictionSubjectMembershipDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.PredictionSubjectID != "" {
		fmt.Fprintf(out, "Prediction Subject: %s\n", details.PredictionSubjectID)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User: %s\n", details.UserID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	fmt.Fprintf(out, "Can Manage Memberships: %t\n", details.CanManageMemberships)
	fmt.Fprintf(out, "Can See Predictions Without Creating Prediction: %t\n", details.CanSeePredictionsWithoutCreatingPrediction)
	fmt.Fprintf(out, "Can Update Prediction Consensus: %t\n", details.CanUpdatePredictionConsensus)
	fmt.Fprintf(out, "Can Update or Destroy Others' Predictions: %t\n", details.CanUpdateOrDestroyOthersPredictions)
	fmt.Fprintf(out, "Can Manage Gaps: %t\n", details.CanManageGaps)

	return nil
}
