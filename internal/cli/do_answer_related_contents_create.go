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

type doAnswerRelatedContentsCreateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	Answer             string
	RelatedContentType string
	RelatedContentID   string
}

func newDoAnswerRelatedContentsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an answer related content link",
		Long: `Create an answer related content link.

Required flags:
  --answer                Answer ID (required)
  --related-content-type  Related content type (required)
  --related-content-id    Related content ID (required)

Related content types:
  newsletters, glossary-terms, release-notes, press-releases, objectives,
  features, questions (also accepts class names like GlossaryTerm)`,
		Example: `  # Link an answer to a newsletter
  xbe do answer-related-contents create \
    --answer 123 \
    --related-content-type newsletters \
    --related-content-id 456`,
		Args: cobra.NoArgs,
		RunE: runDoAnswerRelatedContentsCreate,
	}
	initDoAnswerRelatedContentsCreateFlags(cmd)
	return cmd
}

func init() {
	doAnswerRelatedContentsCmd.AddCommand(newDoAnswerRelatedContentsCreateCmd())
}

func initDoAnswerRelatedContentsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("answer", "", "Answer ID (required)")
	cmd.Flags().String("related-content-type", "", "Related content type (required)")
	cmd.Flags().String("related-content-id", "", "Related content ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoAnswerRelatedContentsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoAnswerRelatedContentsCreateOptions(cmd)
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

	if opts.Answer == "" {
		err := fmt.Errorf("--answer is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.RelatedContentType == "" {
		err := fmt.Errorf("--related-content-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.RelatedContentID == "" {
		err := fmt.Errorf("--related-content-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relatedContentType := normalizeRelatedContentTypeForRelationship(opts.RelatedContentType)

	relationships := map[string]any{
		"answer": map[string]any{
			"data": map[string]any{
				"type": "answers",
				"id":   opts.Answer,
			},
		},
		"related-content": map[string]any{
			"data": map[string]any{
				"type": relatedContentType,
				"id":   opts.RelatedContentID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "answer-related-contents",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/answer-related-contents", jsonBody)
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

	row := answerRelatedContentRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created answer related content %s\n", row.ID)
	return nil
}

func parseDoAnswerRelatedContentsCreateOptions(cmd *cobra.Command) (doAnswerRelatedContentsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	answer, _ := cmd.Flags().GetString("answer")
	relatedContentType, _ := cmd.Flags().GetString("related-content-type")
	relatedContentID, _ := cmd.Flags().GetString("related-content-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doAnswerRelatedContentsCreateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		Answer:             answer,
		RelatedContentType: relatedContentType,
		RelatedContentID:   relatedContentID,
	}, nil
}

func answerRelatedContentRowFromSingle(resp jsonAPISingleResponse) answerRelatedContentRow {
	attrs := resp.Data.Attributes
	row := answerRelatedContentRow{
		ID:         resp.Data.ID,
		Similarity: floatAttr(attrs, "similarity"),
		CreatedAt:  formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:  formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["answer"]; ok && rel.Data != nil {
		row.AnswerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["related-content"]; ok && rel.Data != nil {
		row.RelatedContentType = rel.Data.Type
		row.RelatedContentID = rel.Data.ID
	}

	return row
}
