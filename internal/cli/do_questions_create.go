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

type doQuestionsCreateOptions struct {
	BaseURL                             string
	Token                               string
	JSON                                bool
	Content                             string
	Source                              string
	IgnoreOrganizationScopedNewsletters string
	IsPublic                            string
	IsPublicToOrganizationChildren      string
	AskedBy                             string
	PublicOrganizationScope             string
	PublicOrganizationScopeType         string
	PublicOrganizationScopeID           string
}

func newDoQuestionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a question",
		Long: `Create a question.

Required flags:
  --content  Question content

Optional flags:
  --source                             Source (app/link)
  --ignore-organization-scoped-newsletters  Ignore org-scoped newsletters (true/false)
  --is-public                          Make question public (true/false)
  --is-public-to-organization-children  Share with org children (true/false)
  --asked-by                           Asking user ID
  --public-organization-scope          Public scope (type|id, e.g., brokers|123)
  --public-organization-scope-type     Public scope type (e.g., brokers)
  --public-organization-scope-id       Public scope ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a question
  xbe do questions create --content "What are today's safety priorities?"

  # Create a public question scoped to a broker
  xbe do questions create \
    --content "What changed in the schedule?" \
    --is-public true \
    --public-organization-scope brokers|123

  # Output as JSON
  xbe do questions create --content "How do I reset my password?" --json`,
		Args: cobra.NoArgs,
		RunE: runDoQuestionsCreate,
	}
	initDoQuestionsCreateFlags(cmd)
	return cmd
}

func init() {
	doQuestionsCmd.AddCommand(newDoQuestionsCreateCmd())
}

func initDoQuestionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("content", "", "Question content (required)")
	cmd.Flags().String("source", "", "Source (app/link)")
	cmd.Flags().String("ignore-organization-scoped-newsletters", "", "Ignore org-scoped newsletters (true/false)")
	cmd.Flags().String("is-public", "", "Make question public (true/false)")
	cmd.Flags().String("is-public-to-organization-children", "", "Share with org children (true/false)")
	cmd.Flags().String("asked-by", "", "Asking user ID")
	cmd.Flags().String("public-organization-scope", "", "Public scope (type|id, e.g., brokers|123)")
	cmd.Flags().String("public-organization-scope-type", "", "Public scope type (e.g., brokers)")
	cmd.Flags().String("public-organization-scope-id", "", "Public scope ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoQuestionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoQuestionsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Content) == "" {
		err := fmt.Errorf("--content is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	scopeType, scopeID, err := parsePublicOrganizationScope(opts.PublicOrganizationScope, opts.PublicOrganizationScopeType, opts.PublicOrganizationScopeID)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"content": opts.Content,
	}
	if strings.TrimSpace(opts.Source) != "" {
		attributes["source"] = opts.Source
	}
	if opts.IgnoreOrganizationScopedNewsletters != "" {
		attributes["ignore-organization-scoped-newsletters"] = opts.IgnoreOrganizationScopedNewsletters == "true"
	}
	if opts.IsPublic != "" {
		attributes["is-public"] = opts.IsPublic == "true"
	}
	if opts.IsPublicToOrganizationChildren != "" {
		attributes["is-public-to-organization-children"] = opts.IsPublicToOrganizationChildren == "true"
	}

	relationships := map[string]any{}
	if strings.TrimSpace(opts.AskedBy) != "" {
		relationships["asked-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.AskedBy,
			},
		}
	}
	if scopeType != "" && scopeID != "" {
		relationships["public-organization-scope"] = map[string]any{
			"data": map[string]any{
				"type": scopeType,
				"id":   scopeID,
			},
		}
	}

	requestData := map[string]any{
		"type":       "questions",
		"attributes": attributes,
	}
	if len(relationships) > 0 {
		requestData["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": requestData,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/questions", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created question %s\n", row.ID)
	return nil
}

func parseDoQuestionsCreateOptions(cmd *cobra.Command) (doQuestionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	content, _ := cmd.Flags().GetString("content")
	source, _ := cmd.Flags().GetString("source")
	ignoreOrganizationScopedNewsletters, _ := cmd.Flags().GetString("ignore-organization-scoped-newsletters")
	isPublic, _ := cmd.Flags().GetString("is-public")
	isPublicToOrganizationChildren, _ := cmd.Flags().GetString("is-public-to-organization-children")
	askedBy, _ := cmd.Flags().GetString("asked-by")
	publicOrganizationScope, _ := cmd.Flags().GetString("public-organization-scope")
	publicOrganizationScopeType, _ := cmd.Flags().GetString("public-organization-scope-type")
	publicOrganizationScopeID, _ := cmd.Flags().GetString("public-organization-scope-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doQuestionsCreateOptions{
		BaseURL:                             baseURL,
		Token:                               token,
		JSON:                                jsonOut,
		Content:                             content,
		Source:                              source,
		IgnoreOrganizationScopedNewsletters: ignoreOrganizationScopedNewsletters,
		IsPublic:                            isPublic,
		IsPublicToOrganizationChildren:      isPublicToOrganizationChildren,
		AskedBy:                             askedBy,
		PublicOrganizationScope:             publicOrganizationScope,
		PublicOrganizationScopeType:         publicOrganizationScopeType,
		PublicOrganizationScopeID:           publicOrganizationScopeID,
	}, nil
}
