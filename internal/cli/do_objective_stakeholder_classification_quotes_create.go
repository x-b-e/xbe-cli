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

type doObjectiveStakeholderClassificationQuotesCreateOptions struct {
	BaseURL                            string
	Token                              string
	JSON                               bool
	ObjectiveStakeholderClassification string
	Content                            string
	IsGenerated                        bool
}

func newDoObjectiveStakeholderClassificationQuotesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an objective stakeholder classification quote",
		Long: `Create an objective stakeholder classification quote.

Required flags:
  --objective-stakeholder-classification  Objective stakeholder classification ID
  --content                               Quote content

Optional flags:
  --is-generated  Mark the quote as generated

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a quote
  xbe do objective-stakeholder-classification-quotes create \
    --objective-stakeholder-classification 123 \
    --content "Stakeholder values transparency"

  # Create a generated quote
  xbe do objective-stakeholder-classification-quotes create \
    --objective-stakeholder-classification 123 \
    --content "Generated summary" \
    --is-generated

  # JSON output
  xbe do objective-stakeholder-classification-quotes create \
    --objective-stakeholder-classification 123 \
    --content "Generated summary" \
    --is-generated --json`,
		Args: cobra.NoArgs,
		RunE: runDoObjectiveStakeholderClassificationQuotesCreate,
	}
	initDoObjectiveStakeholderClassificationQuotesCreateFlags(cmd)
	return cmd
}

func init() {
	doObjectiveStakeholderClassificationQuotesCmd.AddCommand(newDoObjectiveStakeholderClassificationQuotesCreateCmd())
}

func initDoObjectiveStakeholderClassificationQuotesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("objective-stakeholder-classification", "", "Objective stakeholder classification ID (required)")
	cmd.Flags().String("content", "", "Quote content (required)")
	cmd.Flags().Bool("is-generated", false, "Mark the quote as generated")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoObjectiveStakeholderClassificationQuotesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoObjectiveStakeholderClassificationQuotesCreateOptions(cmd)
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

	if opts.ObjectiveStakeholderClassification == "" {
		err := fmt.Errorf("--objective-stakeholder-classification is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Content == "" {
		err := fmt.Errorf("--content is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"content": opts.Content,
	}
	if cmd.Flags().Changed("is-generated") {
		attributes["is-generated"] = opts.IsGenerated
	}

	relationships := map[string]any{
		"objective-stakeholder-classification": map[string]any{
			"data": map[string]any{
				"type": "objective-stakeholder-classifications",
				"id":   opts.ObjectiveStakeholderClassification,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "objective-stakeholder-classification-quotes",
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

	body, _, err := client.Post(cmd.Context(), "/v1/objective-stakeholder-classification-quotes", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created objective stakeholder classification quote %s\n", row.ID)
	return nil
}

func parseDoObjectiveStakeholderClassificationQuotesCreateOptions(cmd *cobra.Command) (doObjectiveStakeholderClassificationQuotesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	objectiveStakeholderClassification, _ := cmd.Flags().GetString("objective-stakeholder-classification")
	content, _ := cmd.Flags().GetString("content")
	isGenerated, _ := cmd.Flags().GetBool("is-generated")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doObjectiveStakeholderClassificationQuotesCreateOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		ObjectiveStakeholderClassification: objectiveStakeholderClassification,
		Content:                            content,
		IsGenerated:                        isGenerated,
	}, nil
}
