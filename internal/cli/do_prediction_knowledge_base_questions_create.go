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

type doPredictionKnowledgeBaseQuestionsCreateOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	Title                   string
	Description             string
	Status                  string
	PredictionKnowledgeBase string
	PredictionSubject       string
	CreatedBy               string
}

func newDoPredictionKnowledgeBaseQuestionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a prediction knowledge base question",
		Long: `Create a prediction knowledge base question.

Required flags:
  --prediction-knowledge-base  Prediction knowledge base ID
  --title                      Question title

Optional flags:
  --description                Question description
  --status                     Question status (open/resolved/dismissed)
  --prediction-subject         Prediction subject ID
  --created-by                 Creator user ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a knowledge base question
  xbe do prediction-knowledge-base-questions create \
    --prediction-knowledge-base 123 \
    --title "What are the main risks?" \
    --description "Focus on schedule and safety risks."

  # Create with status
  xbe do prediction-knowledge-base-questions create \
    --prediction-knowledge-base 123 \
    --title "Summarize key assumptions" \
    --status open

  # Output as JSON
  xbe do prediction-knowledge-base-questions create \
    --prediction-knowledge-base 123 \
    --title "What changed since last week?" \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoPredictionKnowledgeBaseQuestionsCreate,
	}
	initDoPredictionKnowledgeBaseQuestionsCreateFlags(cmd)
	return cmd
}

func init() {
	doPredictionKnowledgeBaseQuestionsCmd.AddCommand(newDoPredictionKnowledgeBaseQuestionsCreateCmd())
}

func initDoPredictionKnowledgeBaseQuestionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("prediction-knowledge-base", "", "Prediction knowledge base ID (required)")
	cmd.Flags().String("title", "", "Question title (required)")
	cmd.Flags().String("description", "", "Question description")
	cmd.Flags().String("status", "", "Question status (open/resolved/dismissed)")
	cmd.Flags().String("prediction-subject", "", "Prediction subject ID")
	cmd.Flags().String("created-by", "", "Creator user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPredictionKnowledgeBaseQuestionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoPredictionKnowledgeBaseQuestionsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.PredictionKnowledgeBase) == "" {
		err := fmt.Errorf("--prediction-knowledge-base is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Title) == "" {
		err := fmt.Errorf("--title is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"title": opts.Title,
	}
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}

	relationships := map[string]any{
		"prediction-knowledge-base": map[string]any{
			"data": map[string]any{
				"type": "prediction-knowledge-bases",
				"id":   opts.PredictionKnowledgeBase,
			},
		},
	}

	if strings.TrimSpace(opts.PredictionSubject) != "" {
		relationships["prediction-subject"] = map[string]any{
			"data": map[string]any{
				"type": "prediction-subjects",
				"id":   opts.PredictionSubject,
			},
		}
	}
	if strings.TrimSpace(opts.CreatedBy) != "" {
		relationships["created-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.CreatedBy,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "prediction-knowledge-base-questions",
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

	body, _, err := client.Post(cmd.Context(), "/v1/prediction-knowledge-base-questions", jsonBody)
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

	row := predictionKnowledgeBaseQuestionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created prediction knowledge base question %s\n", row.ID)
	return nil
}

func parseDoPredictionKnowledgeBaseQuestionsCreateOptions(cmd *cobra.Command) (doPredictionKnowledgeBaseQuestionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	predictionKnowledgeBase, _ := cmd.Flags().GetString("prediction-knowledge-base")
	title, _ := cmd.Flags().GetString("title")
	description, _ := cmd.Flags().GetString("description")
	status, _ := cmd.Flags().GetString("status")
	predictionSubject, _ := cmd.Flags().GetString("prediction-subject")
	createdBy, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPredictionKnowledgeBaseQuestionsCreateOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		Title:                   title,
		Description:             description,
		Status:                  status,
		PredictionKnowledgeBase: predictionKnowledgeBase,
		PredictionSubject:       predictionSubject,
		CreatedBy:               createdBy,
	}, nil
}
