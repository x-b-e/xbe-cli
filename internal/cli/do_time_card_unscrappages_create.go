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

type doTimeCardUnscrappagesCreateOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	TimeCard string
	Comment  string
}

func newDoTimeCardUnscrappagesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a time card unscrappage",
		Long: `Create a time card unscrappage.

Required flags:
  --time-card   Time card ID (required)

Optional flags:
  --comment     Comment explaining the change`,
		Example: `  # Unscrap a time card
  xbe do time-card-unscrappages create --time-card 123

  # Unscrap a time card with a comment
  xbe do time-card-unscrappages create --time-card 123 --comment "Restored after review"

  # JSON output
  xbe do time-card-unscrappages create --time-card 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoTimeCardUnscrappagesCreate,
	}
	initDoTimeCardUnscrappagesCreateFlags(cmd)
	return cmd
}

func init() {
	doTimeCardUnscrappagesCmd.AddCommand(newDoTimeCardUnscrappagesCreateCmd())
}

func initDoTimeCardUnscrappagesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("time-card", "", "Time card ID (required)")
	cmd.Flags().String("comment", "", "Comment explaining the change")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTimeCardUnscrappagesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTimeCardUnscrappagesCreateOptions(cmd)
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

	if strings.TrimSpace(opts.TimeCard) == "" {
		err := fmt.Errorf("--time-card is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if strings.TrimSpace(opts.Comment) != "" {
		attributes["comment"] = opts.Comment
	}

	relationships := map[string]any{
		"time-card": map[string]any{
			"data": map[string]any{
				"type": "time-cards",
				"id":   opts.TimeCard,
			},
		},
	}

	data := map[string]any{
		"type":          "time-card-unscrappages",
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

	body, _, err := client.Post(cmd.Context(), "/v1/time-card-unscrappages", jsonBody)
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

	row := buildTimeCardUnscrappageRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created time card unscrappage %s\n", row.ID)
	return nil
}

func parseDoTimeCardUnscrappagesCreateOptions(cmd *cobra.Command) (doTimeCardUnscrappagesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	timeCard, _ := cmd.Flags().GetString("time-card")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeCardUnscrappagesCreateOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		TimeCard: timeCard,
		Comment:  comment,
	}, nil
}
