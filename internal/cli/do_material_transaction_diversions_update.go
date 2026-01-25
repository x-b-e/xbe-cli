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

type doMaterialTransactionDiversionsUpdateOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	ID                   string
	NewJobSite           string
	NewDeliveryDate      string
	DivertedTonsExplicit string
	DriverInstructions   string
}

func newDoMaterialTransactionDiversionsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a material transaction diversion",
		Long: `Update a material transaction diversion.

All flags are optional. Only provided flags will update the diversion.

Optional flags:
  --new-job-site            New job site ID (empty to clear)
  --new-delivery-date       New delivery date (YYYY-MM-DD)
  --diverted-tons-explicit  Explicit diverted tons
  --driver-instructions     Driver instructions`,
		Example: `  # Update driver instructions
  xbe do material-transaction-diversions update 123 --driver-instructions "Call dispatch"

  # Update delivery date and diverted tons
  xbe do material-transaction-diversions update 123 --new-delivery-date 2025-01-03 --diverted-tons-explicit 10.5`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaterialTransactionDiversionsUpdate,
	}
	initDoMaterialTransactionDiversionsUpdateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTransactionDiversionsCmd.AddCommand(newDoMaterialTransactionDiversionsUpdateCmd())
}

func initDoMaterialTransactionDiversionsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("new-job-site", "", "New job site ID (empty to clear)")
	cmd.Flags().String("new-delivery-date", "", "New delivery date (YYYY-MM-DD)")
	cmd.Flags().String("diverted-tons-explicit", "", "Explicit diverted tons")
	cmd.Flags().String("driver-instructions", "", "Driver instructions")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialTransactionDiversionsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaterialTransactionDiversionsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("new-delivery-date") {
		attributes["new-delivery-date"] = opts.NewDeliveryDate
	}
	if cmd.Flags().Changed("diverted-tons-explicit") {
		attributes["diverted-tons-explicit"] = opts.DivertedTonsExplicit
	}
	if cmd.Flags().Changed("driver-instructions") {
		attributes["driver-instructions"] = opts.DriverInstructions
	}

	if cmd.Flags().Changed("new-job-site") {
		if opts.NewJobSite == "" {
			relationships["new-job-site"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["new-job-site"] = map[string]any{
				"data": map[string]any{
					"type": "job-sites",
					"id":   opts.NewJobSite,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("at least one field must be specified for update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "material-transaction-diversions",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/material-transaction-diversions/"+opts.ID, jsonBody)
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

	row := materialTransactionDiversionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated material transaction diversion %s\n", row.ID)
	return nil
}

func parseDoMaterialTransactionDiversionsUpdateOptions(cmd *cobra.Command, args []string) (doMaterialTransactionDiversionsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	newJobSite, _ := cmd.Flags().GetString("new-job-site")
	newDeliveryDate, _ := cmd.Flags().GetString("new-delivery-date")
	divertedTonsExplicit, _ := cmd.Flags().GetString("diverted-tons-explicit")
	driverInstructions, _ := cmd.Flags().GetString("driver-instructions")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTransactionDiversionsUpdateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		ID:                   args[0],
		NewJobSite:           newJobSite,
		NewDeliveryDate:      newDeliveryDate,
		DivertedTonsExplicit: divertedTonsExplicit,
		DriverInstructions:   driverInstructions,
	}, nil
}
