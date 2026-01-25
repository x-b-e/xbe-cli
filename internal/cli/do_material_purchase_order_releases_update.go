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

type doMaterialPurchaseOrderReleasesUpdateOptions struct {
	BaseURL                            string
	Token                              string
	JSON                               bool
	ID                                 string
	Status                             string
	Quantity                           string
	Trucker                            string
	TenderJobScheduleShift             string
	JobScheduleShift                   string
	SkipValidateTenderJobShiftMatch    bool
	SkipValidateTenderJobShiftMatchSet bool
}

func newDoMaterialPurchaseOrderReleasesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a material purchase order release",
		Long: `Update a material purchase order release.

Writable attributes:
  --status        Release status (editing,approved,redeemed,closed)
  --quantity      Release quantity
  --skip-validate-tender-job-schedule-shift-match  Skip tender shift match validation

Relationships:
  --trucker                 Trucker ID
  --tender-job-schedule-shift  Tender job schedule shift ID
  --job-schedule-shift       Job schedule shift ID`,
		Example: `  # Update quantity
  xbe do material-purchase-order-releases update 123 --quantity 12

  # Update status
  xbe do material-purchase-order-releases update 123 --status approved`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaterialPurchaseOrderReleasesUpdate,
	}
	initDoMaterialPurchaseOrderReleasesUpdateFlags(cmd)
	return cmd
}

func init() {
	doMaterialPurchaseOrderReleasesCmd.AddCommand(newDoMaterialPurchaseOrderReleasesUpdateCmd())
}

func initDoMaterialPurchaseOrderReleasesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("status", "", "Release status (editing,approved,redeemed,closed)")
	cmd.Flags().String("quantity", "", "Release quantity")
	cmd.Flags().String("trucker", "", "Trucker ID")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID")
	cmd.Flags().String("job-schedule-shift", "", "Job schedule shift ID")
	cmd.Flags().Bool("skip-validate-tender-job-schedule-shift-match", false, "Skip tender shift match validation")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialPurchaseOrderReleasesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaterialPurchaseOrderReleasesUpdateOptions(cmd, args)
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
	if opts.Status != "" {
		attributes["status"] = opts.Status
	}
	if opts.Quantity != "" {
		attributes["quantity"] = opts.Quantity
	}
	if opts.SkipValidateTenderJobShiftMatchSet {
		attributes["skip-validate-tender-job-schedule-shift-match"] = opts.SkipValidateTenderJobShiftMatch
	}

	relationships := map[string]any{}
	if opts.Trucker != "" {
		relationships["trucker"] = map[string]any{
			"data": map[string]any{
				"type": "truckers",
				"id":   opts.Trucker,
			},
		}
	}
	if opts.TenderJobScheduleShift != "" {
		relationships["tender-job-schedule-shift"] = map[string]any{
			"data": map[string]any{
				"type": "tender-job-schedule-shifts",
				"id":   opts.TenderJobScheduleShift,
			},
		}
	}
	if opts.JobScheduleShift != "" {
		relationships["job-schedule-shift"] = map[string]any{
			"data": map[string]any{
				"type": "job-schedule-shifts",
				"id":   opts.JobScheduleShift,
			},
		}
	}

	data := map[string]any{
		"type": "material-purchase-order-releases",
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

	path := fmt.Sprintf("/v1/material-purchase-order-releases/%s", opts.ID)
	body, _, err := client.Patch(cmd.Context(), path, jsonBody)
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

	row := materialPurchaseOrderReleaseRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated material purchase order release %s\n", row.ID)
	return nil
}

func parseDoMaterialPurchaseOrderReleasesUpdateOptions(cmd *cobra.Command, args []string) (doMaterialPurchaseOrderReleasesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	status, _ := cmd.Flags().GetString("status")
	quantity, _ := cmd.Flags().GetString("quantity")
	trucker, _ := cmd.Flags().GetString("trucker")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	jobScheduleShift, _ := cmd.Flags().GetString("job-schedule-shift")
	skipValidate, _ := cmd.Flags().GetBool("skip-validate-tender-job-schedule-shift-match")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialPurchaseOrderReleasesUpdateOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		ID:                                 args[0],
		Status:                             status,
		Quantity:                           quantity,
		Trucker:                            trucker,
		TenderJobScheduleShift:             tenderJobScheduleShift,
		JobScheduleShift:                   jobScheduleShift,
		SkipValidateTenderJobShiftMatch:    skipValidate,
		SkipValidateTenderJobShiftMatchSet: cmd.Flags().Changed("skip-validate-tender-job-schedule-shift-match"),
	}, nil
}
