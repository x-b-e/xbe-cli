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

type doReactionClassificationsCreateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	CreatedBy string
}

func newDoReactionClassificationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new reaction classification",
		Long: `Create a new reaction classification.

Note: Reaction classifications have read-only label, utf8, and external_reference
attributes. Only the created_by attribute can be set on create.

Optional flags:
  --created-by    User ID of the creator`,
		Example: `  # Create a reaction classification
  xbe do reaction-classifications create --created-by 123`,
		RunE: runDoReactionClassificationsCreate,
	}
	initDoReactionClassificationsCreateFlags(cmd)
	return cmd
}

func init() {
	doReactionClassificationsCmd.AddCommand(newDoReactionClassificationsCreateCmd())
}

func initDoReactionClassificationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("created-by", "", "User ID of the creator")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoReactionClassificationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoReactionClassificationsCreateOptions(cmd)
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

	if opts.CreatedBy != "" {
		attributes["created-by"] = opts.CreatedBy
	}

	data := map[string]any{
		"type":       "reaction-classifications",
		"attributes": attributes,
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/reaction-classifications", jsonBody)
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
		return writeJSON(cmd.OutOrStdout(), map[string]string{
			"id":    resp.Data.ID,
			"label": stringAttr(resp.Data.Attributes, "label"),
		})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created reaction classification %s\n", resp.Data.ID)
	return nil
}

func parseDoReactionClassificationsCreateOptions(cmd *cobra.Command) (doReactionClassificationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	createdBy, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doReactionClassificationsCreateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		CreatedBy: createdBy,
	}, nil
}
