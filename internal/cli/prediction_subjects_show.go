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

type predictionSubjectsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type predictionSubjectDetails struct {
	ID                                       string   `json:"id"`
	Name                                     string   `json:"name,omitempty"`
	Description                              string   `json:"description,omitempty"`
	Status                                   string   `json:"status,omitempty"`
	Kind                                     string   `json:"kind,omitempty"`
	Actual                                   string   `json:"actual,omitempty"`
	ActualDueAt                              string   `json:"actual_due_at,omitempty"`
	PredictionsDueAt                         string   `json:"predictions_due_at,omitempty"`
	DomainMin                                string   `json:"domain_min,omitempty"`
	DomainMax                                string   `json:"domain_max,omitempty"`
	ApprovedExplainedGapPct                  string   `json:"approved_explained_gap_pct,omitempty"`
	AdditionalAttributes                     any      `json:"additional_attributes,omitempty"`
	ReferenceNumber                          any      `json:"reference_number,omitempty"`
	BrokerID                                 string   `json:"broker_id,omitempty"`
	ParentType                               string   `json:"parent_type,omitempty"`
	ParentID                                 string   `json:"parent_id,omitempty"`
	CreatedByID                              string   `json:"created_by_id,omitempty"`
	PredictionConsensusID                    string   `json:"prediction_consensus_id,omitempty"`
	BusinessUnitID                           string   `json:"business_unit_id,omitempty"`
	LowestLosingBidPredictionSubjectDetailID string   `json:"lowest_losing_bid_prediction_subject_detail_id,omitempty"`
	PredictionIDs                            []string `json:"prediction_ids,omitempty"`
	PredictionSubjectMembershipIDs           []string `json:"prediction_subject_membership_ids,omitempty"`
	PredictionSubjectGapIDs                  []string `json:"prediction_subject_gap_ids,omitempty"`
	PredictionKnowledgeBaseQuestionIDs       []string `json:"prediction_knowledge_base_question_ids,omitempty"`
	PredictionAgentIDs                       []string `json:"prediction_agent_ids,omitempty"`
	RecapID                                  string   `json:"recap_id,omitempty"`
	TaggingIDs                               []string `json:"tagging_ids,omitempty"`
	TagIDs                                   []string `json:"tag_ids,omitempty"`
}

func newPredictionSubjectsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show prediction subject details",
		Long: `Show the full details of a prediction subject.

Output Fields:
  ID                                       Prediction subject identifier
  Name                                     Subject name
  Description                              Subject description
  Status                                   Current status
  Kind                                     Prediction kind
  Actual                                   Actual value (if set)
  Actual Due At                            When the actual value is due
  Predictions Due At                       When predictions are due
  Domain Min                               Domain minimum value
  Domain Max                               Domain maximum value
  Approved Explained Gap %                 Approved explained gap percentage
  Additional Attributes                    Additional attributes JSON
  Reference Number                         Reference number
  Broker ID                                Associated broker ID
  Parent                                   Parent type and ID
  Created By ID                            Creator user ID
  Prediction Consensus ID                  Prediction consensus ID
  Business Unit ID                         Associated business unit ID
  Lowest Losing Bid Prediction Subject Detail ID  Detail ID (if available)
  Prediction IDs                           Related prediction IDs
  Prediction Subject Membership IDs        Related membership IDs
  Prediction Subject Gap IDs               Related gap IDs
  Prediction Knowledge Base Question IDs   Related knowledge base question IDs
  Prediction Agent IDs                     Related prediction agent IDs
  Recap ID                                 Latest recap ID
  Tagging IDs                              Related tagging IDs
  Tag IDs                                  Related tag IDs

Arguments:
  <id>    The prediction subject ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a prediction subject
  xbe view prediction-subjects show 123

  # JSON output
  xbe view prediction-subjects show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runPredictionSubjectsShow,
	}
	initPredictionSubjectsShowFlags(cmd)
	return cmd
}

func init() {
	predictionSubjectsCmd.AddCommand(newPredictionSubjectsShowCmd())
}

func initPredictionSubjectsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPredictionSubjectsShow(cmd *cobra.Command, args []string) error {
	opts, err := parsePredictionSubjectsShowOptions(cmd)
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
		return fmt.Errorf("prediction subject id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "broker,parent,created-by,lowest-losing-bid-prediction-subject-detail,prediction-consensus,business-unit,predictions,prediction-subject-memberships,prediction-subject-gaps,prediction-knowledge-base-questions,prediction-agents,recap,taggings,tags")

	body, _, err := client.Get(cmd.Context(), "/v1/prediction-subjects/"+id, query)
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

	details := buildPredictionSubjectDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderPredictionSubjectDetails(cmd, details)
}

func parsePredictionSubjectsShowOptions(cmd *cobra.Command) (predictionSubjectsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return predictionSubjectsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildPredictionSubjectDetails(resp jsonAPISingleResponse) predictionSubjectDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := predictionSubjectDetails{
		ID:                      resource.ID,
		Name:                    stringAttr(attrs, "name"),
		Description:             stringAttr(attrs, "description"),
		Status:                  stringAttr(attrs, "status"),
		Kind:                    stringAttr(attrs, "kind"),
		Actual:                  stringAttr(attrs, "actual"),
		ActualDueAt:             stringAttr(attrs, "actual-due-at"),
		PredictionsDueAt:        stringAttr(attrs, "predictions-due-at"),
		DomainMin:               stringAttr(attrs, "domain-min"),
		DomainMax:               stringAttr(attrs, "domain-max"),
		ApprovedExplainedGapPct: stringAttr(attrs, "approved-explained-gap-pct"),
	}

	if value, ok := attrs["additional-attributes"]; ok {
		details.AdditionalAttributes = value
	}
	if value, ok := attrs["reference-number"]; ok {
		details.ReferenceNumber = value
	}

	if rel, ok := resource.Relationships["parent"]; ok && rel.Data != nil {
		details.ParentType = rel.Data.Type
		details.ParentID = rel.Data.ID
	}

	details.BrokerID = relationshipIDFromMap(resource.Relationships, "broker")
	details.CreatedByID = relationshipIDFromMap(resource.Relationships, "created-by")
	details.PredictionConsensusID = relationshipIDFromMap(resource.Relationships, "prediction-consensus")
	details.BusinessUnitID = relationshipIDFromMap(resource.Relationships, "business-unit")
	details.LowestLosingBidPredictionSubjectDetailID = relationshipIDFromMap(resource.Relationships, "lowest-losing-bid-prediction-subject-detail")
	details.PredictionIDs = relationshipIDsFromMap(resource.Relationships, "predictions")
	details.PredictionSubjectMembershipIDs = relationshipIDsFromMap(resource.Relationships, "prediction-subject-memberships")
	details.PredictionSubjectGapIDs = relationshipIDsFromMap(resource.Relationships, "prediction-subject-gaps")
	details.PredictionKnowledgeBaseQuestionIDs = relationshipIDsFromMap(resource.Relationships, "prediction-knowledge-base-questions")
	details.PredictionAgentIDs = relationshipIDsFromMap(resource.Relationships, "prediction-agents")
	details.RecapID = relationshipIDFromMap(resource.Relationships, "recap")
	details.TaggingIDs = relationshipIDsFromMap(resource.Relationships, "taggings")
	details.TagIDs = relationshipIDsFromMap(resource.Relationships, "tags")

	return details
}

func renderPredictionSubjectDetails(cmd *cobra.Command, details predictionSubjectDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}
	if details.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", details.Description)
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.Kind != "" {
		fmt.Fprintf(out, "Kind: %s\n", details.Kind)
	}
	if details.Actual != "" {
		fmt.Fprintf(out, "Actual: %s\n", details.Actual)
	}
	if details.ActualDueAt != "" {
		fmt.Fprintf(out, "Actual Due At: %s\n", details.ActualDueAt)
	}
	if details.PredictionsDueAt != "" {
		fmt.Fprintf(out, "Predictions Due At: %s\n", details.PredictionsDueAt)
	}
	if details.DomainMin != "" {
		fmt.Fprintf(out, "Domain Min: %s\n", details.DomainMin)
	}
	if details.DomainMax != "" {
		fmt.Fprintf(out, "Domain Max: %s\n", details.DomainMax)
	}
	if details.ApprovedExplainedGapPct != "" {
		fmt.Fprintf(out, "Approved Explained Gap %%: %s\n", details.ApprovedExplainedGapPct)
	}
	if details.AdditionalAttributes != nil {
		fmt.Fprintln(out, "Additional Attributes:")
		fmt.Fprintln(out, formatJSONBlock(details.AdditionalAttributes, "  "))
	}
	if details.ReferenceNumber != nil {
		fmt.Fprintln(out, "Reference Number:")
		fmt.Fprintln(out, formatJSONBlock(details.ReferenceNumber, "  "))
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.ParentType != "" && details.ParentID != "" {
		fmt.Fprintf(out, "Parent: %s/%s\n", details.ParentType, details.ParentID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By ID: %s\n", details.CreatedByID)
	}
	if details.PredictionConsensusID != "" {
		fmt.Fprintf(out, "Prediction Consensus ID: %s\n", details.PredictionConsensusID)
	}
	if details.BusinessUnitID != "" {
		fmt.Fprintf(out, "Business Unit ID: %s\n", details.BusinessUnitID)
	}
	if details.LowestLosingBidPredictionSubjectDetailID != "" {
		fmt.Fprintf(out, "Lowest Losing Bid Prediction Subject Detail ID: %s\n", details.LowestLosingBidPredictionSubjectDetailID)
	}
	if len(details.PredictionIDs) > 0 {
		fmt.Fprintf(out, "Prediction IDs: %s\n", strings.Join(details.PredictionIDs, ", "))
	}
	if len(details.PredictionSubjectMembershipIDs) > 0 {
		fmt.Fprintf(out, "Prediction Subject Membership IDs: %s\n", strings.Join(details.PredictionSubjectMembershipIDs, ", "))
	}
	if len(details.PredictionSubjectGapIDs) > 0 {
		fmt.Fprintf(out, "Prediction Subject Gap IDs: %s\n", strings.Join(details.PredictionSubjectGapIDs, ", "))
	}
	if len(details.PredictionKnowledgeBaseQuestionIDs) > 0 {
		fmt.Fprintf(out, "Prediction Knowledge Base Question IDs: %s\n", strings.Join(details.PredictionKnowledgeBaseQuestionIDs, ", "))
	}
	if len(details.PredictionAgentIDs) > 0 {
		fmt.Fprintf(out, "Prediction Agent IDs: %s\n", strings.Join(details.PredictionAgentIDs, ", "))
	}
	if details.RecapID != "" {
		fmt.Fprintf(out, "Recap ID: %s\n", details.RecapID)
	}
	if len(details.TaggingIDs) > 0 {
		fmt.Fprintf(out, "Tagging IDs: %s\n", strings.Join(details.TaggingIDs, ", "))
	}
	if len(details.TagIDs) > 0 {
		fmt.Fprintf(out, "Tag IDs: %s\n", strings.Join(details.TagIDs, ", "))
	}

	return nil
}
