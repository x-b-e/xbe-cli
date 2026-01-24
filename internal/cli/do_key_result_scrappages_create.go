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

type doKeyResultScrappagesCreateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	KeyResult string
	Comment   string
}

func newDoKeyResultScrappagesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a key result scrappage",
		Long: `Create a key result scrappage.

Required flags:
  --key-result   Key result ID (required)

Optional flags:
  --comment      Comment explaining the change`,
		Example: `  # Scrap a key result
  xbe do key-result-scrappages create --key-result 123

  # Scrap a key result with a comment
  xbe do key-result-scrappages create --key-result 123 --comment "Archived"

  # JSON output
  xbe do key-result-scrappages create --key-result 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoKeyResultScrappagesCreate,
	}
	initDoKeyResultScrappagesCreateFlags(cmd)
	return cmd
}

func init() {
	doKeyResultScrappagesCmd.AddCommand(newDoKeyResultScrappagesCreateCmd())
}

func initDoKeyResultScrappagesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("key-result", "", "Key result ID (required)")
	cmd.Flags().String("comment", "", "Comment explaining the change")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoKeyResultScrappagesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoKeyResultScrappagesCreateOptions(cmd)
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

	if strings.TrimSpace(opts.KeyResult) == "" {
		err := fmt.Errorf("--key-result is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if strings.TrimSpace(opts.Comment) != "" {
		attributes["comment"] = opts.Comment
	}

	relationships := map[string]any{
		"key-result": map[string]any{
			"data": map[string]any{
				"type": "key-results",
				"id":   opts.KeyResult,
			},
		},
	}

	data := map[string]any{
		"type":          "key-result-scrappages",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/key-result-scrappages", jsonBody)
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

	row := buildKeyResultScrappageRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created key result scrappage %s\n", row.ID)
	return nil
}

func parseDoKeyResultScrappagesCreateOptions(cmd *cobra.Command) (doKeyResultScrappagesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	keyResult, _ := cmd.Flags().GetString("key-result")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doKeyResultScrappagesCreateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		KeyResult: keyResult,
		Comment:   comment,
	}, nil
}
