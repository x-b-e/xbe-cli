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

type doTruckerBrokeragesUpdateOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	ID              string
	Trucker         string
	BrokeredTrucker string
}

func newDoTruckerBrokeragesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a trucker brokerage",
		Long: `Update an existing trucker brokerage.

Arguments:
  <id>    The trucker brokerage ID (required)

Flags:
  --trucker           Brokering trucker ID
  --brokered-trucker  Brokered trucker ID`,
		Example: `  # Update a trucker brokerage
  xbe do trucker-brokerages update 123 --trucker 456 --brokered-trucker 789

  # Get JSON output
  xbe do trucker-brokerages update 123 --brokered-trucker 789 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTruckerBrokeragesUpdate,
	}
	initDoTruckerBrokeragesUpdateFlags(cmd)
	return cmd
}

func init() {
	doTruckerBrokeragesCmd.AddCommand(newDoTruckerBrokeragesUpdateCmd())
}

func initDoTruckerBrokeragesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("trucker", "", "Brokering trucker ID")
	cmd.Flags().String("brokered-trucker", "", "Brokered trucker ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTruckerBrokeragesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTruckerBrokeragesUpdateOptions(cmd, args)
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

	relationships := map[string]any{}

	if cmd.Flags().Changed("trucker") {
		if opts.Trucker == "" {
			relationships["trucker"] = map[string]any{"data": nil}
		} else {
			relationships["trucker"] = map[string]any{
				"data": map[string]any{
					"type": "truckers",
					"id":   opts.Trucker,
				},
			}
		}
	}
	if cmd.Flags().Changed("brokered-trucker") {
		if opts.BrokeredTrucker == "" {
			relationships["brokered-trucker"] = map[string]any{"data": nil}
		} else {
			relationships["brokered-trucker"] = map[string]any{
				"data": map[string]any{
					"type": "truckers",
					"id":   opts.BrokeredTrucker,
				},
			}
		}
	}

	if len(relationships) == 0 {
		err := fmt.Errorf("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":          "trucker-brokerages",
		"id":            opts.ID,
		"relationships": relationships,
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/trucker-brokerages/"+opts.ID, jsonBody)
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

	details := buildTruckerBrokerageDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated trucker brokerage %s\n", details.ID)
	return nil
}

func parseDoTruckerBrokeragesUpdateOptions(cmd *cobra.Command, args []string) (doTruckerBrokeragesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	trucker, _ := cmd.Flags().GetString("trucker")
	brokeredTrucker, _ := cmd.Flags().GetString("brokered-trucker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTruckerBrokeragesUpdateOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		ID:              args[0],
		Trucker:         trucker,
		BrokeredTrucker: brokeredTrucker,
	}, nil
}
