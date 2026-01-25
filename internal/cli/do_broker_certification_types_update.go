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

type doBrokerCertificationTypesUpdateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	ID                  string
	BrokerID            string
	CertificationTypeID string
}

func newDoBrokerCertificationTypesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a broker certification type",
		Long: `Update a broker certification type.

Arguments:
  <id>  The broker certification type ID (required)

Optional flags:
  --broker               Broker ID
  --certification-type   Certification type ID

Note: The certification type must belong to the broker.`,
		Example: `  # Update a broker certification type
  xbe do broker-certification-types update 123 --broker 456 --certification-type 789`,
		Args: cobra.ExactArgs(1),
		RunE: runDoBrokerCertificationTypesUpdate,
	}
	initDoBrokerCertificationTypesUpdateFlags(cmd)
	return cmd
}

func init() {
	doBrokerCertificationTypesCmd.AddCommand(newDoBrokerCertificationTypesUpdateCmd())
}

func initDoBrokerCertificationTypesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("broker", "", "Broker ID")
	cmd.Flags().String("certification-type", "", "Certification type ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBrokerCertificationTypesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoBrokerCertificationTypesUpdateOptions(cmd, args)
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
		return fmt.Errorf("broker certification type id is required")
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

	if cmd.Flags().Changed("certification-type") {
		if strings.TrimSpace(opts.CertificationTypeID) == "" {
			return fmt.Errorf("--certification-type cannot be empty")
		}
		relationships["certification-type"] = map[string]any{
			"data": map[string]any{
				"type": "certification-types",
				"id":   opts.CertificationTypeID,
			},
		}
	}

	if len(relationships) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":          "broker-certification-types",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/broker-certification-types/"+opts.ID, jsonBody)
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

	row := brokerCertificationTypeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated broker certification type %s\n", row.ID)
	return nil
}

func parseDoBrokerCertificationTypesUpdateOptions(cmd *cobra.Command, args []string) (doBrokerCertificationTypesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	brokerID, _ := cmd.Flags().GetString("broker")
	certificationTypeID, _ := cmd.Flags().GetString("certification-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBrokerCertificationTypesUpdateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		ID:                  args[0],
		BrokerID:            brokerID,
		CertificationTypeID: certificationTypeID,
	}, nil
}
