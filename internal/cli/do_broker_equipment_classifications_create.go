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

type doBrokerEquipmentClassificationsCreateOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	BrokerID                  string
	EquipmentClassificationID string
}

func newDoBrokerEquipmentClassificationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a broker equipment classification",
		Long: `Create a broker equipment classification.

Required flags:
  --broker                   Broker ID (required)
  --equipment-classification Equipment classification ID (required)

Note: The equipment classification must be a non-root classification (has a parent).`,
		Example: `  # Create a broker equipment classification
  xbe do broker-equipment-classifications create --broker 123 --equipment-classification 456`,
		Args: cobra.NoArgs,
		RunE: runDoBrokerEquipmentClassificationsCreate,
	}
	initDoBrokerEquipmentClassificationsCreateFlags(cmd)
	return cmd
}

func init() {
	doBrokerEquipmentClassificationsCmd.AddCommand(newDoBrokerEquipmentClassificationsCreateCmd())
}

func initDoBrokerEquipmentClassificationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("equipment-classification", "", "Equipment classification ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("broker")
	cmd.MarkFlagRequired("equipment-classification")
}

func runDoBrokerEquipmentClassificationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoBrokerEquipmentClassificationsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.BrokerID) == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.EquipmentClassificationID) == "" {
		err := fmt.Errorf("--equipment-classification is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.BrokerID,
			},
		},
		"equipment-classification": map[string]any{
			"data": map[string]any{
				"type": "equipment-classifications",
				"id":   opts.EquipmentClassificationID,
			},
		},
	}

	data := map[string]any{
		"type":          "broker-equipment-classifications",
		"relationships": relationships,
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/broker-equipment-classifications", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created broker equipment classification %s\n", row.ID)
	return nil
}

func parseDoBrokerEquipmentClassificationsCreateOptions(cmd *cobra.Command) (doBrokerEquipmentClassificationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	brokerID, _ := cmd.Flags().GetString("broker")
	equipmentClassificationID, _ := cmd.Flags().GetString("equipment-classification")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBrokerEquipmentClassificationsCreateOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		BrokerID:                  brokerID,
		EquipmentClassificationID: equipmentClassificationID,
	}, nil
}
