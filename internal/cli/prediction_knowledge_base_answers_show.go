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

type predictionKnowledgeBaseAnswersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type predictionKnowledgeBaseAnswerDetails struct {
	ID                                string `json:"id"`
	Content                           string `json:"content,omitempty"`
	PredictionKnowledgeBaseQuestionID string `json:"prediction_knowledge_base_question_id,omitempty"`
}

func newPredictionKnowledgeBaseAnswersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show prediction knowledge base answer details",
		Long: `Show the full details of a prediction knowledge base answer.

Output Fields:
  ID           Answer identifier
  QUESTION ID  Prediction knowledge base question ID
  CONTENT      Answer content

Arguments:
  <id>    The prediction knowledge base answer ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --no-auth, --base-url, --token`,
		Example: `  # Show a prediction knowledge base answer
  xbe view prediction-knowledge-base-answers show 123

  # Output as JSON
  xbe view prediction-knowledge-base-answers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runPredictionKnowledgeBaseAnswersShow,
	}
	initPredictionKnowledgeBaseAnswersShowFlags(cmd)
	return cmd
}

func init() {
	predictionKnowledgeBaseAnswersCmd.AddCommand(newPredictionKnowledgeBaseAnswersShowCmd())
}

func initPredictionKnowledgeBaseAnswersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPredictionKnowledgeBaseAnswersShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parsePredictionKnowledgeBaseAnswersShowOptions(cmd)
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
		return fmt.Errorf("prediction knowledge base answer id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[prediction-knowledge-base-answers]", "content,prediction-knowledge-base-question")

	body, _, err := client.Get(cmd.Context(), "/v1/prediction-knowledge-base-answers/"+id, query)
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

	details := buildPredictionKnowledgeBaseAnswerDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderPredictionKnowledgeBaseAnswerDetails(cmd, details)
}

func parsePredictionKnowledgeBaseAnswersShowOptions(cmd *cobra.Command) (predictionKnowledgeBaseAnswersShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return predictionKnowledgeBaseAnswersShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return predictionKnowledgeBaseAnswersShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return predictionKnowledgeBaseAnswersShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return predictionKnowledgeBaseAnswersShowOptions{}, err
	}

	return predictionKnowledgeBaseAnswersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildPredictionKnowledgeBaseAnswerDetails(resp jsonAPISingleResponse) predictionKnowledgeBaseAnswerDetails {
	resource := resp.Data
	attrs := resource.Attributes
	return predictionKnowledgeBaseAnswerDetails{
		ID:                                resource.ID,
		Content:                           strings.TrimSpace(stringAttr(attrs, "content")),
		PredictionKnowledgeBaseQuestionID: relationshipIDFromMap(resource.Relationships, "prediction-knowledge-base-question"),
	}
}

func renderPredictionKnowledgeBaseAnswerDetails(cmd *cobra.Command, details predictionKnowledgeBaseAnswerDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.PredictionKnowledgeBaseQuestionID != "" {
		fmt.Fprintf(out, "Prediction Knowledge Base Question ID: %s\n", details.PredictionKnowledgeBaseQuestionID)
	}
	if details.Content != "" {
		fmt.Fprintf(out, "Content: %s\n", details.Content)
	}

	return nil
}
