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

type doCommentsCreateOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	Body            string
	CommentableType string
	CommentableID   string
	DoNotNotify     bool
	IncludeInRecap  bool
	IsAdminOnly     bool
}

func newDoCommentsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new comment",
		Long: `Create a new comment on a resource.

Required flags:
  --body                Body text (required)
  --commentable-type    Type of resource to comment on (required, e.g., projects, truckers)
  --commentable-id      ID of resource to comment on (required)

Optional flags:
  --do-not-notify       Do not send notifications
  --include-in-recap    Include in recap
  --is-admin-only       Mark as admin-only comment`,
		Example: `  # Create a comment on a project
  xbe do comments create \
    --body "This is a comment" \
    --commentable-type projects \
    --commentable-id 123

  # Create a comment without notifications
  xbe do comments create \
    --body "Silent comment" \
    --commentable-type truckers \
    --commentable-id 456 \
    --do-not-notify

  # Get JSON output
  xbe do comments create \
    --body "Comment text" \
    --commentable-type projects \
    --commentable-id 123 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoCommentsCreate,
	}
	initDoCommentsCreateFlags(cmd)
	return cmd
}

func init() {
	doCommentsCmd.AddCommand(newDoCommentsCreateCmd())
}

func initDoCommentsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("body", "", "Comment body text (required)")
	cmd.Flags().String("commentable-type", "", "Type of resource to comment on (required)")
	cmd.Flags().String("commentable-id", "", "ID of resource to comment on (required)")
	cmd.Flags().Bool("do-not-notify", false, "Do not send notifications")
	cmd.Flags().Bool("include-in-recap", false, "Include in recap")
	cmd.Flags().Bool("is-admin-only", false, "Mark as admin-only comment")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCommentsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCommentsCreateOptions(cmd)
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

	if opts.Body == "" {
		err := fmt.Errorf("--body is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.CommentableType == "" {
		err := fmt.Errorf("--commentable-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.CommentableID == "" {
		err := fmt.Errorf("--commentable-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"body": opts.Body,
	}

	if cmd.Flags().Changed("do-not-notify") {
		attributes["do-not-notify"] = opts.DoNotNotify
	}
	if cmd.Flags().Changed("include-in-recap") {
		attributes["include-in-recap"] = opts.IncludeInRecap
	}
	if cmd.Flags().Changed("is-admin-only") {
		attributes["is-admin-only"] = opts.IsAdminOnly
	}

	relationships := map[string]any{
		"commentable": map[string]any{
			"data": map[string]any{
				"type": opts.CommentableType,
				"id":   opts.CommentableID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "comments",
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

	body, _, err := client.Post(cmd.Context(), "/v1/comments", jsonBody)
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

	row := buildCommentRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created comment %s\n", row.ID)
	return nil
}

func parseDoCommentsCreateOptions(cmd *cobra.Command) (doCommentsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	body, _ := cmd.Flags().GetString("body")
	commentableType, _ := cmd.Flags().GetString("commentable-type")
	commentableID, _ := cmd.Flags().GetString("commentable-id")
	doNotNotify, _ := cmd.Flags().GetBool("do-not-notify")
	includeInRecap, _ := cmd.Flags().GetBool("include-in-recap")
	isAdminOnly, _ := cmd.Flags().GetBool("is-admin-only")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCommentsCreateOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		Body:            body,
		CommentableType: commentableType,
		CommentableID:   commentableID,
		DoNotNotify:     doNotNotify,
		IncludeInRecap:  includeInRecap,
		IsAdminOnly:     isAdminOnly,
	}, nil
}

func buildCommentRowFromSingle(resp jsonAPISingleResponse) commentRow {
	attrs := resp.Data.Attributes

	row := commentRow{
		ID:             resp.Data.ID,
		Body:           stringAttr(attrs, "body"),
		IsAdminOnly:    boolAttr(attrs, "is-admin-only"),
		DoNotNotify:    boolAttr(attrs, "do-not-notify"),
		IncludeInRecap: boolAttr(attrs, "include-in-recap"),
		IsCreatedByBot: boolAttr(attrs, "is-created-by-bot"),
	}

	if rel, ok := resp.Data.Relationships["commentable"]; ok && rel.Data != nil {
		row.CommentableType = rel.Data.Type
		row.CommentableID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	return row
}
