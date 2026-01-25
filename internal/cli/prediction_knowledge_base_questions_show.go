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

type predictionKnowledgeBaseQuestionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type predictionKnowledgeBaseQuestionDetails struct {
	ID                        string   `json:"id"`
	Title                     string   `json:"title,omitempty"`
	Description               string   `json:"description,omitempty"`
	Status                    string   `json:"status,omitempty"`
	PredictionKnowledgeBaseID string   `json:"prediction_knowledge_base_id,omitempty"`
	PredictionSubjectID       string   `json:"prediction_subject_id,omitempty"`
	CreatedByID               string   `json:"created_by_id,omitempty"`
	AnswerID                  string   `json:"answer_id,omitempty"`
	TagIDs                    []string `json:"tag_ids,omitempty"`
	TaggingIDs                []string `json:"tagging_ids,omitempty"`
}

func newPredictionKnowledgeBaseQuestionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show prediction knowledge base question details",
		Long: `Show the full details of a prediction knowledge base question.

Output Fields:
  ID                    Question identifier
  Title                 Question title
  Description           Question description
  Status                Question status
  Prediction Knowledge Base  Knowledge base ID
  Prediction Subject    Prediction subject ID (if present)
  Created By            Creator user ID
  Answer                Answer ID (if present)
  Tags                  Tag IDs (if present)
  Taggings              Tagging IDs (if present)

Arguments:
  <id>  The prediction knowledge base question ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show prediction knowledge base question details
  xbe view prediction-knowledge-base-questions show 123

  # Output as JSON
  xbe view prediction-knowledge-base-questions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runPredictionKnowledgeBaseQuestionsShow,
	}
	initPredictionKnowledgeBaseQuestionsShowFlags(cmd)
	return cmd
}

func init() {
	predictionKnowledgeBaseQuestionsCmd.AddCommand(newPredictionKnowledgeBaseQuestionsShowCmd())
}

func initPredictionKnowledgeBaseQuestionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPredictionKnowledgeBaseQuestionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parsePredictionKnowledgeBaseQuestionsShowOptions(cmd)
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
		return fmt.Errorf("prediction knowledge base question id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[prediction-knowledge-base-questions]", "title,description,status,prediction-subject,prediction-knowledge-base,created-by,answer,tags,taggings")

	body, _, err := client.Get(cmd.Context(), "/v1/prediction-knowledge-base-questions/"+id, query)
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

	details := buildPredictionKnowledgeBaseQuestionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderPredictionKnowledgeBaseQuestionDetails(cmd, details)
}

func parsePredictionKnowledgeBaseQuestionsShowOptions(cmd *cobra.Command) (predictionKnowledgeBaseQuestionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return predictionKnowledgeBaseQuestionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildPredictionKnowledgeBaseQuestionDetails(resp jsonAPISingleResponse) predictionKnowledgeBaseQuestionDetails {
	attrs := resp.Data.Attributes
	details := predictionKnowledgeBaseQuestionDetails{
		ID:          resp.Data.ID,
		Title:       stringAttr(attrs, "title"),
		Description: stringAttr(attrs, "description"),
		Status:      stringAttr(attrs, "status"),
	}

	if rel, ok := resp.Data.Relationships["prediction-knowledge-base"]; ok && rel.Data != nil {
		details.PredictionKnowledgeBaseID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["prediction-subject"]; ok && rel.Data != nil {
		details.PredictionSubjectID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["answer"]; ok && rel.Data != nil {
		details.AnswerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["tags"]; ok {
		details.TagIDs = relationshipIDsToStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["taggings"]; ok {
		details.TaggingIDs = relationshipIDsToStrings(rel)
	}

	return details
}

func renderPredictionKnowledgeBaseQuestionDetails(cmd *cobra.Command, details predictionKnowledgeBaseQuestionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Title != "" {
		fmt.Fprintf(out, "Title: %s\n", details.Title)
	}
	if details.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", details.Description)
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.PredictionKnowledgeBaseID != "" {
		fmt.Fprintf(out, "Prediction Knowledge Base: %s\n", details.PredictionKnowledgeBaseID)
	}
	if details.PredictionSubjectID != "" {
		fmt.Fprintf(out, "Prediction Subject: %s\n", details.PredictionSubjectID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	if details.AnswerID != "" {
		fmt.Fprintf(out, "Answer: %s\n", details.AnswerID)
	}
	if len(details.TagIDs) > 0 {
		fmt.Fprintf(out, "Tags: %s\n", strings.Join(details.TagIDs, ", "))
	}
	if len(details.TaggingIDs) > 0 {
		fmt.Fprintf(out, "Taggings: %s\n", strings.Join(details.TaggingIDs, ", "))
	}

	return nil
}
