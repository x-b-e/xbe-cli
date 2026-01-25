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

type doRetainerDeductionsUpdateOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	ID       string
	Retainer string
	Amount   string
	Note     string
}

func newDoRetainerDeductionsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a retainer deduction",
		Long: `Update a retainer deduction.

Optional flags:
  --amount    Deduction amount
  --note      Note for the deduction
  --retainer  Retainer ID`,
		Example: `  # Update a retainer deduction amount
  xbe do retainer-deductions update 123 --amount 125.75

  # Update note
  xbe do retainer-deductions update 123 --note "Updated note"

  # Update retainer relationship
  xbe do retainer-deductions update 123 --retainer 456`,
		Args: cobra.ExactArgs(1),
		RunE: runDoRetainerDeductionsUpdate,
	}
	initDoRetainerDeductionsUpdateFlags(cmd)
	return cmd
}

func init() {
	doRetainerDeductionsCmd.AddCommand(newDoRetainerDeductionsUpdateCmd())
}

func initDoRetainerDeductionsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("amount", "", "Deduction amount")
	cmd.Flags().String("note", "", "Note for the deduction")
	cmd.Flags().String("retainer", "", "Retainer ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoRetainerDeductionsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoRetainerDeductionsUpdateOptions(cmd, args)
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
	relationships := map[string]any{}

	if cmd.Flags().Changed("amount") {
		attributes["amount"] = opts.Amount
	}
	if cmd.Flags().Changed("note") {
		attributes["note"] = opts.Note
	}
	if cmd.Flags().Changed("retainer") {
		relationships["retainer"] = map[string]any{
			"data": map[string]any{
				"type": "retainers",
				"id":   opts.Retainer,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "retainer-deductions",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
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

	body, _, err := client.Patch(cmd.Context(), "/v1/retainer-deductions/"+opts.ID, jsonBody)
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
		row := buildRetainerDeductionRow(resp.Data)
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated retainer deduction %s\n", resp.Data.ID)
	return nil
}

func parseDoRetainerDeductionsUpdateOptions(cmd *cobra.Command, args []string) (doRetainerDeductionsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	amount, _ := cmd.Flags().GetString("amount")
	note, _ := cmd.Flags().GetString("note")
	retainer, _ := cmd.Flags().GetString("retainer")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRetainerDeductionsUpdateOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		ID:       args[0],
		Retainer: retainer,
		Amount:   amount,
		Note:     note,
	}, nil
}
