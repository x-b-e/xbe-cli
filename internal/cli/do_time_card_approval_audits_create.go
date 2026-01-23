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

type doTimeCardApprovalAuditsCreateOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	TimeCardID string
	UserID     string
	Note       string
}

func newDoTimeCardApprovalAuditsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a time card approval audit",
		Long: `Create a time card approval audit.

Required flags:
  --time-card  Approved time card ID (required)
  --user       User ID (required)

Optional flags:
  --note       Audit note`,
		Example: `  # Create a time card approval audit
  xbe do time-card-approval-audits create --time-card 123 --user 456 --note "Reviewed"

  # Create without a note
  xbe do time-card-approval-audits create --time-card 123 --user 456

  # Output as JSON
  xbe do time-card-approval-audits create --time-card 123 --user 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoTimeCardApprovalAuditsCreate,
	}
	initDoTimeCardApprovalAuditsCreateFlags(cmd)
	return cmd
}

func init() {
	doTimeCardApprovalAuditsCmd.AddCommand(newDoTimeCardApprovalAuditsCreateCmd())
}

func initDoTimeCardApprovalAuditsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("time-card", "", "Approved time card ID (required)")
	cmd.Flags().String("user", "", "User ID (required)")
	cmd.Flags().String("note", "", "Audit note")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("time-card")
	_ = cmd.MarkFlagRequired("user")
}

func runDoTimeCardApprovalAuditsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTimeCardApprovalAuditsCreateOptions(cmd)
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

	if opts.TimeCardID == "" {
		err := fmt.Errorf("--time-card is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.UserID == "" {
		err := fmt.Errorf("--user is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.Note != "" {
		attributes["note"] = opts.Note
	}

	relationships := map[string]any{
		"time-card": map[string]any{
			"data": map[string]any{
				"type": "time-cards",
				"id":   opts.TimeCardID,
			},
		},
		"user": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.UserID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "time-card-approval-audits",
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

	body, _, err := client.Post(cmd.Context(), "/v1/time-card-approval-audits", jsonBody)
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

	row := buildTimeCardApprovalAuditRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created time card approval audit %s\n", row.ID)
	return nil
}

func parseDoTimeCardApprovalAuditsCreateOptions(cmd *cobra.Command) (doTimeCardApprovalAuditsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	timeCardID, _ := cmd.Flags().GetString("time-card")
	userID, _ := cmd.Flags().GetString("user")
	note, _ := cmd.Flags().GetString("note")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeCardApprovalAuditsCreateOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		TimeCardID: timeCardID,
		UserID:     userID,
		Note:       note,
	}, nil
}
