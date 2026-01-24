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

type doBrokerEquipmentClassificationsUpdateOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	ID                        string
	BrokerID                  string
	EquipmentClassificationID string
}

func newDoBrokerEquipmentClassificationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a broker equipment classification",
		Long: `Update a broker equipment classification.

Arguments:
  <id>  The broker equipment classification ID (required)

Optional flags:
  --broker                   Broker ID
  --equipment-classification Equipment classification ID

Note: The equipment classification must be a non-root classification (has a parent).`,
		Example: `  # Update a broker equipment classification
  xbe do broker-equipment-classifications update 123 --broker 456 --equipment-classification 789`,
		Args: cobra.ExactArgs(1),
		RunE: runDoBrokerEquipmentClassificationsUpdate,
	}
	initDoBrokerEquipmentClassificationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doBrokerEquipmentClassificationsCmd.AddCommand(newDoBrokerEquipmentClassificationsUpdateCmd())
}

func initDoBrokerEquipmentClassificationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("broker", "", "Broker ID")
	cmd.Flags().String("equipment-classification", "", "Equipment classification ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBrokerEquipmentClassificationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoBrokerEquipmentClassificationsUpdateOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	if strings.TrimSpace(opts.ID) == "" {
		return fmt.Errorf("broker equipment classification id is required")
	}

	relationships := map[string]any{}

	if cmd.Flags().Changed("broker") {
		if strings.TrimSpace(opts.BrokerID) == "" {
			return fmt.Errorf("--broker cannot be empty")
		}
		relationships["broker"] = map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.BrokerID,
			},
		}
	}

	if cmd.Flags().Changed("equipment-classification") {
		if strings.TrimSpace(opts.EquipmentClassificationID) == "" {
			return fmt.Errorf("--equipment-classification cannot be empty")
		}
		relationships["equipment-classification"] = map[string]any{
			"data": map[string]any{
				"type": "equipment-classifications",
				"id":   opts.EquipmentClassificationID,
			},
		}
	}

	if len(relationships) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":          "broker-equipment-classifications",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/broker-equipment-classifications/"+opts.ID, jsonBody)
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

	row := brokerEquipmentClassificationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated broker equipment classification %s\n", row.ID)
	return nil
}

func parseDoBrokerEquipmentClassificationsUpdateOptions(cmd *cobra.Command, args []string) (doBrokerEquipmentClassificationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	brokerID, _ := cmd.Flags().GetString("broker")
	equipmentClassificationID, _ := cmd.Flags().GetString("equipment-classification")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBrokerEquipmentClassificationsUpdateOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		ID:                        args[0],
		BrokerID:                  brokerID,
		EquipmentClassificationID: equipmentClassificationID,
	}, nil
}
