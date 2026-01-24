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

type doRetainerDeductionsCreateOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	Retainer string
	Amount   string
	Note     string
}

func newDoRetainerDeductionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a retainer deduction",
		Long: `Create a retainer deduction.

Required flags:
  --retainer  Retainer ID
  --amount    Deduction amount

Optional flags:
  --note      Note for the deduction`,
		Example: `  # Create a retainer deduction
  xbe do retainer-deductions create --retainer 123 --amount 100.50

  # Create with a note
  xbe do retainer-deductions create --retainer 123 --amount 250 --note "Fuel surcharge"

  # Output JSON
  xbe do retainer-deductions create --retainer 123 --amount 100.50 --json`,
		Args: cobra.NoArgs,
		RunE: runDoRetainerDeductionsCreate,
	}
	initDoRetainerDeductionsCreateFlags(cmd)
	return cmd
}

func init() {
	doRetainerDeductionsCmd.AddCommand(newDoRetainerDeductionsCreateCmd())
}

func initDoRetainerDeductionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("retainer", "", "Retainer ID (required)")
	cmd.Flags().String("amount", "", "Deduction amount (required)")
	cmd.Flags().String("note", "", "Note for the deduction")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("retainer")
	_ = cmd.MarkFlagRequired("amount")
}

func runDoRetainerDeductionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoRetainerDeductionsCreateOptions(cmd)
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

	attributes := map[string]any{
		"amount": opts.Amount,
	}
	if cmd.Flags().Changed("note") {
		attributes["note"] = opts.Note
	}

	relationships := map[string]any{
		"retainer": map[string]any{
			"data": map[string]any{
				"type": "retainers",
				"id":   opts.Retainer,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "retainer-deductions",
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

	body, _, err := client.Post(cmd.Context(), "/v1/retainer-deductions", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created retainer deduction %s\n", resp.Data.ID)
	return nil
}

func parseDoRetainerDeductionsCreateOptions(cmd *cobra.Command) (doRetainerDeductionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	retainer, _ := cmd.Flags().GetString("retainer")
	amount, _ := cmd.Flags().GetString("amount")
	note, _ := cmd.Flags().GetString("note")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRetainerDeductionsCreateOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		Retainer: retainer,
		Amount:   amount,
		Note:     note,
	}, nil
}
