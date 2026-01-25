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

type doCommentReactionsCreateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	Comment                string
	ReactionClassification string
	CreatedBy              string
}

func newDoCommentReactionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a comment reaction",
		Long: `Create a comment reaction.

Required flags:
  --comment                 Comment ID (required)
  --reaction-classification Reaction classification ID (required)

Optional flags:
  --created-by              Created-by user ID (defaults to current user)

Note: The server sets created-by to the current user if omitted.`,
		Example: `  # Add a reaction to a comment
  xbe do comment-reactions create --comment 123 --reaction-classification 456`,
		Args: cobra.NoArgs,
		RunE: runDoCommentReactionsCreate,
	}
	initDoCommentReactionsCreateFlags(cmd)
	return cmd
}

func init() {
	doCommentReactionsCmd.AddCommand(newDoCommentReactionsCreateCmd())
}

func initDoCommentReactionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("comment", "", "Comment ID (required)")
	cmd.Flags().String("reaction-classification", "", "Reaction classification ID (required)")
	cmd.Flags().String("created-by", "", "Created-by user ID (defaults to current user)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCommentReactionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoCommentReactionsCreateOptions(cmd)
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

	if opts.Comment == "" {
		err := fmt.Errorf("--comment is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.ReactionClassification == "" {
		err := fmt.Errorf("--reaction-classification is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"comment": map[string]any{
			"data": map[string]any{
				"type": "comments",
				"id":   opts.Comment,
			},
		},
		"reaction-classification": map[string]any{
			"data": map[string]any{
				"type": "reaction-classifications",
				"id":   opts.ReactionClassification,
			},
		},
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
			"type":          "comment-reactions",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/comment-reactions", jsonBody)
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

	row := commentReactionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created comment reaction %s\n", row.ID)
	return nil
}

func parseDoCommentReactionsCreateOptions(cmd *cobra.Command) (doCommentReactionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	comment, _ := cmd.Flags().GetString("comment")
	reactionClassification, _ := cmd.Flags().GetString("reaction-classification")
	createdBy, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCommentReactionsCreateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		Comment:                comment,
		ReactionClassification: reactionClassification,
		CreatedBy:              createdBy,
	}, nil
}
