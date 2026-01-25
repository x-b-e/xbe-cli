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

type doQuestionsUpdateOptions struct {
	BaseURL                             string
	Token                               string
	JSON                                bool
	ID                                  string
	Content                             string
	Source                              string
	IgnoreOrganizationScopedNewsletters string
	IsPublic                            string
	IsPublicToOrganizationChildren      string
	Motivation                          string
	IsTriaged                           string
	RecreateAnswer                      string
	AskedBy                             string
	AssignedTo                          string
	PublicOrganizationScope             string
	PublicOrganizationScopeType         string
	PublicOrganizationScopeID           string
}

func newDoQuestionsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a question",
		Long: `Update a question.

Optional flags:
  --content                             Question content
  --source                              Source (app/link)
  --ignore-organization-scoped-newsletters  Ignore org-scoped newsletters (true/false)
  --is-public                           Make question public (true/false)
  --is-public-to-organization-children  Share with org children (true/false)
  --motivation                          Motivation (serious/silly, admin only)
  --is-triaged                          Triaged status (true/false, admin only)
  --recreate-answer                     Recreate answer (true/false)
  --asked-by                            Asking user ID (admin only)
  --assigned-to                         Assigned user ID (empty to clear, admin only)
  --public-organization-scope           Public scope (type|id, empty to clear)
  --public-organization-scope-type      Public scope type (e.g., brokers)
  --public-organization-scope-id        Public scope ID

Notes:
  Some fields require admin permissions.

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update content
  xbe do questions update 123 --content "Updated question text"

  # Mark triaged
  xbe do questions update 123 --is-triaged true

  # Clear public scope
  xbe do questions update 123 --public-organization-scope ""`,
		Args: cobra.ExactArgs(1),
		RunE: runDoQuestionsUpdate,
	}
	initDoQuestionsUpdateFlags(cmd)
	return cmd
}

func init() {
	doQuestionsCmd.AddCommand(newDoQuestionsUpdateCmd())
}

func initDoQuestionsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("content", "", "Question content")
	cmd.Flags().String("source", "", "Source (app/link)")
	cmd.Flags().String("ignore-organization-scoped-newsletters", "", "Ignore org-scoped newsletters (true/false)")
	cmd.Flags().String("is-public", "", "Make question public (true/false)")
	cmd.Flags().String("is-public-to-organization-children", "", "Share with org children (true/false)")
	cmd.Flags().String("motivation", "", "Motivation (serious/silly, admin only)")
	cmd.Flags().String("is-triaged", "", "Triaged status (true/false, admin only)")
	cmd.Flags().String("recreate-answer", "", "Recreate answer (true/false)")
	cmd.Flags().String("asked-by", "", "Asking user ID (admin only)")
	cmd.Flags().String("assigned-to", "", "Assigned user ID (empty to clear, admin only)")
	cmd.Flags().String("public-organization-scope", "", "Public scope (type|id, empty to clear)")
	cmd.Flags().String("public-organization-scope-type", "", "Public scope type (e.g., brokers)")
	cmd.Flags().String("public-organization-scope-id", "", "Public scope ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoQuestionsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoQuestionsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("content") {
		attributes["content"] = opts.Content
	}
	if cmd.Flags().Changed("source") {
		attributes["source"] = opts.Source
	}
	if cmd.Flags().Changed("ignore-organization-scoped-newsletters") {
		attributes["ignore-organization-scoped-newsletters"] = opts.IgnoreOrganizationScopedNewsletters == "true"
	}
	if cmd.Flags().Changed("is-public") {
		attributes["is-public"] = opts.IsPublic == "true"
	}
	if cmd.Flags().Changed("is-public-to-organization-children") {
		attributes["is-public-to-organization-children"] = opts.IsPublicToOrganizationChildren == "true"
	}
	if cmd.Flags().Changed("motivation") {
		attributes["motivation"] = opts.Motivation
	}
	if cmd.Flags().Changed("is-triaged") {
		attributes["is-triaged"] = opts.IsTriaged == "true"
	}
	if cmd.Flags().Changed("recreate-answer") {
		attributes["recreate-answer"] = opts.RecreateAnswer == "true"
	}

	if cmd.Flags().Changed("asked-by") {
		if strings.TrimSpace(opts.AskedBy) == "" {
			err := fmt.Errorf("--asked-by cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["asked-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.AskedBy,
			},
		}
	}

	if cmd.Flags().Changed("assigned-to") {
		if strings.TrimSpace(opts.AssignedTo) == "" {
			relationships["assigned-to"] = map[string]any{"data": nil}
		} else {
			relationships["assigned-to"] = map[string]any{
				"data": map[string]any{
					"type": "users",
					"id":   opts.AssignedTo,
				},
			}
		}
	}

	if cmd.Flags().Changed("public-organization-scope") || cmd.Flags().Changed("public-organization-scope-type") || cmd.Flags().Changed("public-organization-scope-id") {
		scopeType := strings.TrimSpace(opts.PublicOrganizationScopeType)
		scopeID := strings.TrimSpace(opts.PublicOrganizationScopeID)
		scope := strings.TrimSpace(opts.PublicOrganizationScope)

		if scope == "" && scopeType == "" && scopeID == "" {
			relationships["public-organization-scope"] = map[string]any{"data": nil}
		} else {
			parsedType, parsedID, err := parsePublicOrganizationScope(scope, scopeType, scopeID)
			if err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), err)
				return err
			}
			relationships["public-organization-scope"] = map[string]any{
				"data": map[string]any{
					"type": parsedType,
					"id":   parsedID,
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
		"type": "questions",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/questions/"+opts.ID, jsonBody)
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

	row := questionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated question %s\n", row.ID)
	return nil
}

func parseDoQuestionsUpdateOptions(cmd *cobra.Command, args []string) (doQuestionsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	content, _ := cmd.Flags().GetString("content")
	source, _ := cmd.Flags().GetString("source")
	ignoreOrganizationScopedNewsletters, _ := cmd.Flags().GetString("ignore-organization-scoped-newsletters")
	isPublic, _ := cmd.Flags().GetString("is-public")
	isPublicToOrganizationChildren, _ := cmd.Flags().GetString("is-public-to-organization-children")
	motivation, _ := cmd.Flags().GetString("motivation")
	isTriaged, _ := cmd.Flags().GetString("is-triaged")
	recreateAnswer, _ := cmd.Flags().GetString("recreate-answer")
	askedBy, _ := cmd.Flags().GetString("asked-by")
	assignedTo, _ := cmd.Flags().GetString("assigned-to")
	publicOrganizationScope, _ := cmd.Flags().GetString("public-organization-scope")
	publicOrganizationScopeType, _ := cmd.Flags().GetString("public-organization-scope-type")
	publicOrganizationScopeID, _ := cmd.Flags().GetString("public-organization-scope-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doQuestionsUpdateOptions{
		BaseURL:                             baseURL,
		Token:                               token,
		JSON:                                jsonOut,
		ID:                                  args[0],
		Content:                             content,
		Source:                              source,
		IgnoreOrganizationScopedNewsletters: ignoreOrganizationScopedNewsletters,
		IsPublic:                            isPublic,
		IsPublicToOrganizationChildren:      isPublicToOrganizationChildren,
		Motivation:                          motivation,
		IsTriaged:                           isTriaged,
		RecreateAnswer:                      recreateAnswer,
		AskedBy:                             askedBy,
		AssignedTo:                          assignedTo,
		PublicOrganizationScope:             publicOrganizationScope,
		PublicOrganizationScopeType:         publicOrganizationScopeType,
		PublicOrganizationScopeID:           publicOrganizationScopeID,
	}, nil
}
