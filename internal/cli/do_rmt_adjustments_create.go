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

type doRmtAdjustmentsCreateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	RmtIDs             []string
	Note               string
	RawDataAdjustments string
	UpdateIfInvoiced   bool
}

type rmtAdjustmentRow struct {
	ID                 string   `json:"id"`
	RmtIDs             []string `json:"rmt_ids,omitempty"`
	Note               string   `json:"note,omitempty"`
	RawDataAdjustments any      `json:"raw_data_adjustments,omitempty"`
	UpdateIfInvoiced   bool     `json:"update_if_invoiced,omitempty"`
	Results            any      `json:"results,omitempty"`
	Messages           any      `json:"messages,omitempty"`
}

func newDoRmtAdjustmentsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an RMT adjustment",
		Long: `Create an RMT adjustment.

Required flags:
  --rmt-ids               Raw material transaction IDs (required, comma-separated or repeated)
  --note                  Adjustment note (required)
  --raw-data-adjustments  Adjustment data as JSON (required)

Optional flags:
  --update-if-invoiced  Update invoiced RMTs when true (default: false)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Adjust RMTs
  xbe do rmt-adjustments create --rmt-ids 123,456 --note "Corrected weights" \
    --raw-data-adjustments '{"net_weight":12.5,"net_weight_by_xbe_reason":"Scale correction"}'

  # Adjust an invoiced RMT
  xbe do rmt-adjustments create --rmt-ids 123 --note "Voided ticket" \
    --raw-data-adjustments '{"is_voided":true,"is_voided_by_xbe_reason":"Duplicate"}' \
    --update-if-invoiced

  # Output as JSON
  xbe do rmt-adjustments create --rmt-ids 123 --note "Corrected weights" \
    --raw-data-adjustments '{"net_weight":12.5,"net_weight_by_xbe_reason":"Scale correction"}' --json`,
		Args: cobra.NoArgs,
		RunE: runDoRmtAdjustmentsCreate,
	}
	initDoRmtAdjustmentsCreateFlags(cmd)
	return cmd
}

func init() {
	doRmtAdjustmentsCmd.AddCommand(newDoRmtAdjustmentsCreateCmd())
}

func initDoRmtAdjustmentsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().StringSlice("rmt-ids", nil, "Raw material transaction IDs (required, comma-separated or repeated)")
	cmd.Flags().String("note", "", "Adjustment note (required)")
	cmd.Flags().String("raw-data-adjustments", "", "Adjustment data as JSON (required)")
	cmd.Flags().Bool("update-if-invoiced", false, "Update invoiced RMTs when true")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoRmtAdjustmentsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoRmtAdjustmentsCreateOptions(cmd)
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

	rmtIDs := make([]string, 0, len(opts.RmtIDs))
	for _, id := range opts.RmtIDs {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" {
			rmtIDs = append(rmtIDs, trimmed)
		}
	}
	if len(rmtIDs) == 0 {
		err := fmt.Errorf("--rmt-ids is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	note := strings.TrimSpace(opts.Note)
	if note == "" {
		err := fmt.Errorf("--note is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.RawDataAdjustments) == "" {
		err := fmt.Errorf("--raw-data-adjustments is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var rawData any
	if err := json.Unmarshal([]byte(opts.RawDataAdjustments), &rawData); err != nil {
		return fmt.Errorf("invalid raw-data-adjustments JSON: %w", err)
	}
	rawDataMap, ok := rawData.(map[string]any)
	if !ok {
		return fmt.Errorf("raw-data-adjustments must be a JSON object")
	}

	attributes := map[string]any{
		"rmt-ids":              rmtIDs,
		"note":                 note,
		"raw-data-adjustments": rawDataMap,
		"update-if-invoiced":   opts.UpdateIfInvoiced,
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "rmt-adjustments",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/sombreros/rmt-adjustments", jsonBody)
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

	row := rmtAdjustmentRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created rmt adjustment %s\n", row.ID)
	return nil
}

func rmtAdjustmentRowFromSingle(resp jsonAPISingleResponse) rmtAdjustmentRow {
	attrs := resp.Data.Attributes
	row := rmtAdjustmentRow{
		ID:               resp.Data.ID,
		RmtIDs:           stringSliceAttr(attrs, "rmt-ids"),
		Note:             stringAttr(attrs, "note"),
		UpdateIfInvoiced: boolAttr(attrs, "update-if-invoiced"),
	}

	if attrs != nil {
		if value, ok := attrs["raw-data-adjustments"]; ok {
			row.RawDataAdjustments = value
		}
		if value, ok := attrs["results"]; ok {
			row.Results = value
		}
		if value, ok := attrs["messages"]; ok {
			row.Messages = value
		}
	}

	return row
}

func parseDoRmtAdjustmentsCreateOptions(cmd *cobra.Command) (doRmtAdjustmentsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	rmtIDs, _ := cmd.Flags().GetStringSlice("rmt-ids")
	note, _ := cmd.Flags().GetString("note")
	rawDataAdjustments, _ := cmd.Flags().GetString("raw-data-adjustments")
	updateIfInvoiced, _ := cmd.Flags().GetBool("update-if-invoiced")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRmtAdjustmentsCreateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		RmtIDs:             rmtIDs,
		Note:               note,
		RawDataAdjustments: rawDataAdjustments,
		UpdateIfInvoiced:   updateIfInvoiced,
	}, nil
}
