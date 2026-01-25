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

type answerRelatedContentsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type answerRelatedContentDetails struct {
	ID                 string  `json:"id"`
	AnswerID           string  `json:"answer_id,omitempty"`
	RelatedContentType string  `json:"related_content_type,omitempty"`
	RelatedContentID   string  `json:"related_content_id,omitempty"`
	Similarity         float64 `json:"similarity,omitempty"`
	CreatedAt          string  `json:"created_at,omitempty"`
	UpdatedAt          string  `json:"updated_at,omitempty"`
}

func newAnswerRelatedContentsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show answer related content details",
		Long: `Show the details for a specific answer related content link.

Output Fields:
  ID               Related content link identifier
  ANSWER           Answer ID
  RELATED CONTENT  Related content type and ID
  SIMILARITY       Similarity score
  CREATED AT       Creation timestamp
  UPDATED AT       Last update timestamp

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a related content link
  xbe view answer-related-contents show 123

  # Output as JSON
  xbe view answer-related-contents show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runAnswerRelatedContentsShow,
	}
	initAnswerRelatedContentsShowFlags(cmd)
	return cmd
}

func init() {
	answerRelatedContentsCmd.AddCommand(newAnswerRelatedContentsShowCmd())
}

func initAnswerRelatedContentsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runAnswerRelatedContentsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseAnswerRelatedContentsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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
		return fmt.Errorf("answer related content id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[answer-related-contents]", "answer,related-content,similarity,created-at,updated-at")

	body, _, err := client.Get(cmd.Context(), "/v1/answer-related-contents/"+id, query)
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

	details := buildAnswerRelatedContentDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderAnswerRelatedContentDetails(cmd, details)
}

func parseAnswerRelatedContentsShowOptions(cmd *cobra.Command) (answerRelatedContentsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return answerRelatedContentsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildAnswerRelatedContentDetails(resp jsonAPISingleResponse) answerRelatedContentDetails {
	attrs := resp.Data.Attributes
	details := answerRelatedContentDetails{
		ID:         resp.Data.ID,
		Similarity: floatAttr(attrs, "similarity"),
		CreatedAt:  formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:  formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["answer"]; ok && rel.Data != nil {
		details.AnswerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["related-content"]; ok && rel.Data != nil {
		details.RelatedContentType = rel.Data.Type
		details.RelatedContentID = rel.Data.ID
	}

	return details
}

func renderAnswerRelatedContentDetails(cmd *cobra.Command, details answerRelatedContentDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.AnswerID != "" {
		fmt.Fprintf(out, "Answer: %s\n", details.AnswerID)
	}
	if details.RelatedContentType != "" && details.RelatedContentID != "" {
		fmt.Fprintf(out, "Related Content: %s/%s\n", details.RelatedContentType, details.RelatedContentID)
	}
	if details.Similarity != 0 {
		fmt.Fprintf(out, "Similarity: %.4f\n", details.Similarity)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
