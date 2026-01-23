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

type doTimeCardApprovalAuditsUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Note    string
}

func newDoTimeCardApprovalAuditsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a time card approval audit",
		Long: `Update a time card approval audit.

Optional flags:
  --note  Audit note`,
		Example: `  # Update an audit note
  xbe do time-card-approval-audits update 123 --note "Updated note"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTimeCardApprovalAuditsUpdate,
	}
	initDoTimeCardApprovalAuditsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTimeCardApprovalAuditsCmd.AddCommand(newDoTimeCardApprovalAuditsUpdateCmd())
}

func initDoTimeCardApprovalAuditsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("note", "", "Audit note")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTimeCardApprovalAuditsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTimeCardApprovalAuditsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("note") {
		attributes["note"] = opts.Note
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "time-card-approval-audits",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/time-card-approval-audits/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated time card approval audit %s\n", row.ID)
	return nil
}

func parseDoTimeCardApprovalAuditsUpdateOptions(cmd *cobra.Command, args []string) (doTimeCardApprovalAuditsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	note, _ := cmd.Flags().GetString("note")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeCardApprovalAuditsUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Note:    note,
	}, nil
}
