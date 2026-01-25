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

type doObjectiveStakeholderClassificationQuotesUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Content string
}

func newDoObjectiveStakeholderClassificationQuotesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an objective stakeholder classification quote",
		Long: `Update an existing objective stakeholder classification quote.

Optional flags:
  --content  Quote content

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update a quote
  xbe do objective-stakeholder-classification-quotes update 123 --content "Updated content"

  # JSON output
  xbe do objective-stakeholder-classification-quotes update 123 --content "Updated content" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoObjectiveStakeholderClassificationQuotesUpdate,
	}
	initDoObjectiveStakeholderClassificationQuotesUpdateFlags(cmd)
	return cmd
}

func init() {
	doObjectiveStakeholderClassificationQuotesCmd.AddCommand(newDoObjectiveStakeholderClassificationQuotesUpdateCmd())
}

func initDoObjectiveStakeholderClassificationQuotesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("content", "", "Quote content")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoObjectiveStakeholderClassificationQuotesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoObjectiveStakeholderClassificationQuotesUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("content") {
		attributes["content"] = opts.Content
	}

	if len(attributes) == 0 {
		err := errors.New("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":       "objective-stakeholder-classification-quotes",
		"id":         opts.ID,
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

	path := fmt.Sprintf("/v1/objective-stakeholder-classification-quotes/%s", opts.ID)
	body, _, err := client.Patch(cmd.Context(), path, jsonBody)
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

	row := buildObjectiveStakeholderClassificationQuoteRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated objective stakeholder classification quote %s\n", row.ID)
	return nil
}

func parseDoObjectiveStakeholderClassificationQuotesUpdateOptions(cmd *cobra.Command, args []string) (doObjectiveStakeholderClassificationQuotesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	content, _ := cmd.Flags().GetString("content")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doObjectiveStakeholderClassificationQuotesUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Content: content,
	}, nil
}
