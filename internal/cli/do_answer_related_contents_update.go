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

type doAnswerRelatedContentsUpdateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	ID                 string
	Answer             string
	RelatedContentType string
	RelatedContentID   string
}

func newDoAnswerRelatedContentsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an answer related content link",
		Long: `Update an answer related content link.

Optional flags:
  --answer                Answer ID
  --related-content-type  Related content type
  --related-content-id    Related content ID

Related content types:
  newsletters, glossary-terms, release-notes, press-releases, objectives,
  features, questions (also accepts class names like GlossaryTerm)`,
		Example: `  # Update the related content for a link
  xbe do answer-related-contents update 123 \
    --related-content-type features \
    --related-content-id 456`,
		Args: cobra.ExactArgs(1),
		RunE: runDoAnswerRelatedContentsUpdate,
	}
	initDoAnswerRelatedContentsUpdateFlags(cmd)
	return cmd
}

func init() {
	doAnswerRelatedContentsCmd.AddCommand(newDoAnswerRelatedContentsUpdateCmd())
}

func initDoAnswerRelatedContentsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("answer", "", "Answer ID")
	cmd.Flags().String("related-content-type", "", "Related content type")
	cmd.Flags().String("related-content-id", "", "Related content ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoAnswerRelatedContentsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoAnswerRelatedContentsUpdateOptions(cmd, args)
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

	relationships := map[string]any{}

	if cmd.Flags().Changed("answer") {
		if opts.Answer == "" {
			err := fmt.Errorf("--answer cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["answer"] = map[string]any{
			"data": map[string]any{
				"type": "answers",
				"id":   opts.Answer,
			},
		}
	}

	relatedTypeChanged := cmd.Flags().Changed("related-content-type")
	relatedIDChanged := cmd.Flags().Changed("related-content-id")
	if relatedTypeChanged || relatedIDChanged {
		if opts.RelatedContentType == "" || opts.RelatedContentID == "" {
			err := fmt.Errorf("--related-content-type and --related-content-id are required together")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relatedContentType := normalizeRelatedContentTypeForRelationship(opts.RelatedContentType)
		relationships["related-content"] = map[string]any{
			"data": map[string]any{
				"type": relatedContentType,
				"id":   opts.RelatedContentID,
			},
		}
	}

	if len(relationships) == 0 {
		err := fmt.Errorf("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "answer-related-contents",
			"id":            opts.ID,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/answer-related-contents/"+opts.ID, jsonBody)
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

	if opts.JSON {
		row := answerRelatedContentRowFromSingle(resp)
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated answer related content %s\n", resp.Data.ID)
	return nil
}

func parseDoAnswerRelatedContentsUpdateOptions(cmd *cobra.Command, args []string) (doAnswerRelatedContentsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	answer, _ := cmd.Flags().GetString("answer")
	relatedContentType, _ := cmd.Flags().GetString("related-content-type")
	relatedContentID, _ := cmd.Flags().GetString("related-content-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doAnswerRelatedContentsUpdateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		ID:                 args[0],
		Answer:             answer,
		RelatedContentType: relatedContentType,
		RelatedContentID:   relatedContentID,
	}, nil
}
