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

type doTimeCardApprovalsCreateOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	TimeCard                  string
	Comment                   string
	SkipQuantityValidation    bool
	CreateZeroPPUMissingRates bool
}

type timeCardApprovalRow struct {
	ID         string `json:"id"`
	TimeCardID string `json:"time_card_id,omitempty"`
	Comment    string `json:"comment,omitempty"`
}

func newDoTimeCardApprovalsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Approve a time card",
		Long: `Approve a time card.

Required flags:
  --time-card  Time card ID (required)

Optional flags:
  --comment                         Approval comment
  --skip-quantity-validation        Skip quantity validation checks
  --create-zero-ppu-missing-rates   Create zero PPU missing rates`,
		Example: `  # Approve a time card
  xbe do time-card-approvals create --time-card 123 --comment "Approved"

  # Approve and skip quantity validation
  xbe do time-card-approvals create \
    --time-card 123 \
    --skip-quantity-validation \
    --create-zero-ppu-missing-rates`,
		Args: cobra.NoArgs,
		RunE: runDoTimeCardApprovalsCreate,
	}
	initDoTimeCardApprovalsCreateFlags(cmd)
	return cmd
}

func init() {
	doTimeCardApprovalsCmd.AddCommand(newDoTimeCardApprovalsCreateCmd())
}

func initDoTimeCardApprovalsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("time-card", "", "Time card ID (required)")
	cmd.Flags().String("comment", "", "Approval comment")
	cmd.Flags().Bool("skip-quantity-validation", false, "Skip quantity validation checks")
	cmd.Flags().Bool("create-zero-ppu-missing-rates", false, "Create zero PPU missing rates")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTimeCardApprovalsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTimeCardApprovalsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run xbe auth login first.")
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
	if cmd.Flags().Changed("skip-quantity-validation") {
		attributes["skip-quantity-validation"] = opts.SkipQuantityValidation
	}
	if cmd.Flags().Changed("create-zero-ppu-missing-rates") {
		attributes["create-zero-ppu-missing-rates"] = opts.CreateZeroPPUMissingRates
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
			"type":          "time-card-approvals",
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

	body, _, err := client.Post(cmd.Context(), "/v1/time-card-approvals", jsonBody)
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

	row := timeCardApprovalRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created time card approval %s\n", row.ID)
	return nil
}

func timeCardApprovalRowFromSingle(resp jsonAPISingleResponse) timeCardApprovalRow {
	attrs := resp.Data.Attributes
	row := timeCardApprovalRow{
		ID:      resp.Data.ID,
		Comment: stringAttr(attrs, "comment"),
	}

	if rel, ok := resp.Data.Relationships["time-card"]; ok && rel.Data != nil {
		row.TimeCardID = rel.Data.ID
	}

	return row
}

func parseDoTimeCardApprovalsCreateOptions(cmd *cobra.Command) (doTimeCardApprovalsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	timeCard, _ := cmd.Flags().GetString("time-card")
	comment, _ := cmd.Flags().GetString("comment")
	skipQuantityValidation, _ := cmd.Flags().GetBool("skip-quantity-validation")
	createZeroPPUMissingRates, _ := cmd.Flags().GetBool("create-zero-ppu-missing-rates")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeCardApprovalsCreateOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		TimeCard:                  timeCard,
		Comment:                   comment,
		SkipQuantityValidation:    skipQuantityValidation,
		CreateZeroPPUMissingRates: createZeroPPUMissingRates,
	}, nil
}
