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

type doCommentsUpdateOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	ID             string
	Body           string
	DoNotNotify    bool
	IncludeInRecap bool
	IsAdminOnly    bool
}

func newDoCommentsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a comment",
		Long: `Update a comment.

Optional flags:
  --body                Update body text
  --do-not-notify       Update do-not-notify flag
  --include-in-recap    Update include-in-recap flag
  --is-admin-only       Update admin-only flag`,
		Example: `  # Update body
  xbe do comments update 123 --body "Updated comment text"

  # Update flags
  xbe do comments update 123 --include-in-recap`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCommentsUpdate,
	}
	initDoCommentsUpdateFlags(cmd)
	return cmd
}

func init() {
	doCommentsCmd.AddCommand(newDoCommentsUpdateCmd())
}

func initDoCommentsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("body", "", "Comment body text")
	cmd.Flags().Bool("do-not-notify", false, "Do not send notifications")
	cmd.Flags().Bool("include-in-recap", false, "Include in recap")
	cmd.Flags().Bool("is-admin-only", false, "Mark as admin-only comment")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCommentsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCommentsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("body") {
		attributes["body"] = opts.Body
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

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "comments",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/comments/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated comment %s\n", row.ID)
	return nil
}

func parseDoCommentsUpdateOptions(cmd *cobra.Command, args []string) (doCommentsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	body, _ := cmd.Flags().GetString("body")
	doNotNotify, _ := cmd.Flags().GetBool("do-not-notify")
	includeInRecap, _ := cmd.Flags().GetBool("include-in-recap")
	isAdminOnly, _ := cmd.Flags().GetBool("is-admin-only")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCommentsUpdateOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		ID:             args[0],
		Body:           body,
		DoNotNotify:    doNotNotify,
		IncludeInRecap: includeInRecap,
		IsAdminOnly:    isAdminOnly,
	}, nil
}
