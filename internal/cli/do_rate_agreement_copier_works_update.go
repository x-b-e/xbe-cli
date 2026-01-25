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

type doRateAgreementCopierWorksUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Note    string
}

func newDoRateAgreementCopierWorksUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a rate agreement copier work",
		Long: `Update a rate agreement copier work.

Optional flags:
  --note    Update the note`,
		Example: `  # Update the work note
  xbe do rate-agreement-copier-works update 123 --note "Follow-up copy"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoRateAgreementCopierWorksUpdate,
	}
	initDoRateAgreementCopierWorksUpdateFlags(cmd)
	return cmd
}

func init() {
	doRateAgreementCopierWorksCmd.AddCommand(newDoRateAgreementCopierWorksUpdateCmd())
}

func initDoRateAgreementCopierWorksUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("note", "", "Note to set")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoRateAgreementCopierWorksUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoRateAgreementCopierWorksUpdateOptions(cmd, args)
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
			"type":       "rate-agreement-copier-works",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/rate-agreement-copier-works/"+opts.ID, jsonBody)
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

	row := buildRateAgreementCopierWorkRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated rate agreement copier work %s\n", row.ID)
	return nil
}

func parseDoRateAgreementCopierWorksUpdateOptions(cmd *cobra.Command, args []string) (doRateAgreementCopierWorksUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	note, _ := cmd.Flags().GetString("note")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRateAgreementCopierWorksUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Note:    note,
	}, nil
}
