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

type doBrokerCertificationTypesCreateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	BrokerID            string
	CertificationTypeID string
}

func newDoBrokerCertificationTypesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a broker certification type",
		Long: `Create a broker certification type.

Required flags:
  --broker               Broker ID (required)
  --certification-type   Certification type ID (required)

Note: The certification type must belong to the broker.`,
		Example: `  # Create a broker certification type
  xbe do broker-certification-types create --broker 123 --certification-type 456`,
		Args: cobra.NoArgs,
		RunE: runDoBrokerCertificationTypesCreate,
	}
	initDoBrokerCertificationTypesCreateFlags(cmd)
	return cmd
}

func init() {
	doBrokerCertificationTypesCmd.AddCommand(newDoBrokerCertificationTypesCreateCmd())
}

func initDoBrokerCertificationTypesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("certification-type", "", "Certification type ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("broker")
	cmd.MarkFlagRequired("certification-type")
}

func runDoBrokerCertificationTypesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoBrokerCertificationTypesCreateOptions(cmd)
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
	if strings.TrimSpace(opts.CertificationTypeID) == "" {
		err := fmt.Errorf("--certification-type is required")
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
		"certification-type": map[string]any{
			"data": map[string]any{
				"type": "certification-types",
				"id":   opts.CertificationTypeID,
			},
		},
	}

	data := map[string]any{
		"type":          "broker-certification-types",
		"relationships": relationships,
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/broker-certification-types", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created broker certification type %s\n", row.ID)
	return nil
}

func parseDoBrokerCertificationTypesCreateOptions(cmd *cobra.Command) (doBrokerCertificationTypesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	brokerID, _ := cmd.Flags().GetString("broker")
	certificationTypeID, _ := cmd.Flags().GetString("certification-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBrokerCertificationTypesCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		BrokerID:            brokerID,
		CertificationTypeID: certificationTypeID,
	}, nil
}
