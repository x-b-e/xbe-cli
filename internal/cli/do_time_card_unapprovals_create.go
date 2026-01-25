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

type doTimeCardUnapprovalsCreateOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	TimeCard string
	Comment  string
}

func newDoTimeCardUnapprovalsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Unapprove a time card",
		Long: `Unapprove a time card.

Time cards must be approved to be unapproved, returning them to submitted status.
Unapprovals require the same permissions as approvals and cannot be applied when
the time card is already associated with an invoice.

Required flags:
  --time-card   Time card ID

Optional flags:
  --comment     Unapproval comment

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Unapprove a time card with a comment
  xbe do time-card-unapprovals create \
    --time-card 123 \
    --comment "Needs review"

  # Unapprove a time card without a comment
  xbe do time-card-unapprovals create --time-card 123`,
		Args: cobra.NoArgs,
		RunE: runDoTimeCardUnapprovalsCreate,
	}
	initDoTimeCardUnapprovalsCreateFlags(cmd)
	return cmd
}

func init() {
	doTimeCardUnapprovalsCmd.AddCommand(newDoTimeCardUnapprovalsCreateCmd())
}

func initDoTimeCardUnapprovalsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("time-card", "", "Time card ID")
	cmd.Flags().String("comment", "", "Unapproval comment")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTimeCardUnapprovalsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTimeCardUnapprovalsCreateOptions(cmd)
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
			"type":          "time-card-unapprovals",
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

	body, _, err := client.Post(cmd.Context(), "/v1/time-card-unapprovals", jsonBody)
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

	row := buildTimeCardUnapprovalRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created time card unapproval %s\n", row.ID)
	return nil
}

func parseDoTimeCardUnapprovalsCreateOptions(cmd *cobra.Command) (doTimeCardUnapprovalsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	timeCard, _ := cmd.Flags().GetString("time-card")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeCardUnapprovalsCreateOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		TimeCard: timeCard,
		Comment:  comment,
	}, nil
}
