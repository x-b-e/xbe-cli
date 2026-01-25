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

type doPredictionKnowledgeBaseQuestionsUpdateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	ID                string
	Title             string
	Description       string
	Status            string
	PredictionSubject string
}

func newDoPredictionKnowledgeBaseQuestionsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a prediction knowledge base question",
		Long: `Update a prediction knowledge base question.

Optional flags:
  --title                Question title
  --description          Question description
  --status               Question status (open/resolved/dismissed)
  --prediction-subject   Prediction subject ID (empty to clear)

Notes:
  Status updates may be restricted by policy.

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update title and description
  xbe do prediction-knowledge-base-questions update 123 \
    --title "Updated question" \
    --description "Updated context"

  # Update status
  xbe do prediction-knowledge-base-questions update 123 --status resolved

  # Update prediction subject
  xbe do prediction-knowledge-base-questions update 123 --prediction-subject 456`,
		Args: cobra.ExactArgs(1),
		RunE: runDoPredictionKnowledgeBaseQuestionsUpdate,
	}
	initDoPredictionKnowledgeBaseQuestionsUpdateFlags(cmd)
	return cmd
}

func init() {
	doPredictionKnowledgeBaseQuestionsCmd.AddCommand(newDoPredictionKnowledgeBaseQuestionsUpdateCmd())
}

func initDoPredictionKnowledgeBaseQuestionsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("title", "", "Question title")
	cmd.Flags().String("description", "", "Question description")
	cmd.Flags().String("status", "", "Question status (open/resolved/dismissed)")
	cmd.Flags().String("prediction-subject", "", "Prediction subject ID (empty to clear)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPredictionKnowledgeBaseQuestionsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoPredictionKnowledgeBaseQuestionsUpdateOptions(cmd, args)
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

	attributes := map[string]any{}
	relationships := map[string]any{}

	if cmd.Flags().Changed("title") {
		attributes["title"] = opts.Title
	}
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("prediction-subject") {
		if strings.TrimSpace(opts.PredictionSubject) == "" {
			relationships["prediction-subject"] = map[string]any{"data": nil}
		} else {
			relationships["prediction-subject"] = map[string]any{
				"data": map[string]any{
					"type": "prediction-subjects",
					"id":   opts.PredictionSubject,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "prediction-knowledge-base-questions",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/prediction-knowledge-base-questions/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated prediction knowledge base question %s\n", row.ID)
	return nil
}

func parseDoPredictionKnowledgeBaseQuestionsUpdateOptions(cmd *cobra.Command, args []string) (doPredictionKnowledgeBaseQuestionsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	title, _ := cmd.Flags().GetString("title")
	description, _ := cmd.Flags().GetString("description")
	status, _ := cmd.Flags().GetString("status")
	predictionSubject, _ := cmd.Flags().GetString("prediction-subject")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPredictionKnowledgeBaseQuestionsUpdateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		ID:                args[0],
		Title:             title,
		Description:       description,
		Status:            status,
		PredictionSubject: predictionSubject,
	}, nil
}
