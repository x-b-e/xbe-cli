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

type predictionKnowledgeBasesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type predictionKnowledgeBaseDetails struct {
	ID                                 string   `json:"id"`
	BrokerID                           string   `json:"broker_id,omitempty"`
	PredictionKnowledgeBaseQuestionIDs []string `json:"prediction_knowledge_base_question_ids,omitempty"`
	CommentIDs                         []string `json:"comment_ids,omitempty"`
	FileAttachmentIDs                  []string `json:"file_attachment_ids,omitempty"`
}

func newPredictionKnowledgeBasesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show prediction knowledge base details",
		Long: `Show the full details of a prediction knowledge base.

Output Fields:
  ID                                Knowledge base identifier
  Broker ID                         Associated broker ID
  Prediction Knowledge Base Question IDs  Related question IDs
  Comment IDs                       Related comment IDs
  File Attachment IDs               Related file attachment IDs

Arguments:
  <id>    The prediction knowledge base ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a prediction knowledge base
  xbe view prediction-knowledge-bases show 123

  # JSON output
  xbe view prediction-knowledge-bases show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runPredictionKnowledgeBasesShow,
	}
	initPredictionKnowledgeBasesShowFlags(cmd)
	return cmd
}

func init() {
	predictionKnowledgeBasesCmd.AddCommand(newPredictionKnowledgeBasesShowCmd())
}

func initPredictionKnowledgeBasesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPredictionKnowledgeBasesShow(cmd *cobra.Command, args []string) error {
	opts, err := parsePredictionKnowledgeBasesShowOptions(cmd)
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
		return fmt.Errorf("prediction knowledge base id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "broker,prediction-knowledge-base-questions,comments,file-attachments")

	body, _, err := client.Get(cmd.Context(), "/v1/prediction-knowledge-bases/"+id, query)
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

	details := buildPredictionKnowledgeBaseDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderPredictionKnowledgeBaseDetails(cmd, details)
}

func parsePredictionKnowledgeBasesShowOptions(cmd *cobra.Command) (predictionKnowledgeBasesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return predictionKnowledgeBasesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildPredictionKnowledgeBaseDetails(resp jsonAPISingleResponse) predictionKnowledgeBaseDetails {
	resource := resp.Data
	return predictionKnowledgeBaseDetails{
		ID:                                 resource.ID,
		BrokerID:                           relationshipIDFromMap(resource.Relationships, "broker"),
		PredictionKnowledgeBaseQuestionIDs: relationshipIDsFromMap(resource.Relationships, "prediction-knowledge-base-questions"),
		CommentIDs:                         relationshipIDsFromMap(resource.Relationships, "comments"),
		FileAttachmentIDs:                  relationshipIDsFromMap(resource.Relationships, "file-attachments"),
	}
}

func renderPredictionKnowledgeBaseDetails(cmd *cobra.Command, details predictionKnowledgeBaseDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if len(details.PredictionKnowledgeBaseQuestionIDs) > 0 {
		fmt.Fprintf(out, "Prediction Knowledge Base Question IDs: %s\n", strings.Join(details.PredictionKnowledgeBaseQuestionIDs, ", "))
	}
	if len(details.CommentIDs) > 0 {
		fmt.Fprintf(out, "Comment IDs: %s\n", strings.Join(details.CommentIDs, ", "))
	}
	if len(details.FileAttachmentIDs) > 0 {
		fmt.Fprintf(out, "File Attachment IDs: %s\n", strings.Join(details.FileAttachmentIDs, ", "))
	}

	return nil
}
