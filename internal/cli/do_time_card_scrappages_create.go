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

type doTimeCardScrappagesCreateOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	TimeCard string
	Comment  string
}

func newDoTimeCardScrappagesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Scrap a time card",
		Long: `Scrap a time card.

Time cards can be scrapped only when they are in editing or submitted status,
have zero value, and are not already associated with an invoice.

Required flags:
  --time-card   Time card ID

Optional flags:
  --comment     Scrappage comment

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Scrap a time card with a comment
  xbe do time-card-scrappages create \
    --time-card 123 \
    --comment "Zero value time card"

  # Scrap a time card without a comment
  xbe do time-card-scrappages create --time-card 123`,
		Args: cobra.NoArgs,
		RunE: runDoTimeCardScrappagesCreate,
	}
	initDoTimeCardScrappagesCreateFlags(cmd)
	return cmd
}

func init() {
	doTimeCardScrappagesCmd.AddCommand(newDoTimeCardScrappagesCreateCmd())
}

func initDoTimeCardScrappagesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("time-card", "", "Time card ID")
	cmd.Flags().String("comment", "", "Scrappage comment")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTimeCardScrappagesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTimeCardScrappagesCreateOptions(cmd)
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
	if opts.Comment != "" {
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

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "time-card-scrappages",
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

	body, _, err := client.Post(cmd.Context(), "/v1/time-card-scrappages", jsonBody)
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

	row := buildTimeCardScrappageRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created time card scrappage %s\n", row.ID)
	return nil
}

func parseDoTimeCardScrappagesCreateOptions(cmd *cobra.Command) (doTimeCardScrappagesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	timeCard, _ := cmd.Flags().GetString("time-card")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeCardScrappagesCreateOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		TimeCard: timeCard,
		Comment:  comment,
	}, nil
}
