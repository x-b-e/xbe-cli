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

type doWorkOrderServiceCodesCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	Code        string
	Description string
	Broker      string
}

func newDoWorkOrderServiceCodesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new work order service code",
		Long: `Create a new work order service code.

Required flags:
  --code    The service code value (required)
  --broker  The broker ID (required)

Optional flags:
  --description  Description of the service code`,
		Example: `  # Create a work order service code
  xbe do work-order-service-codes create --code "HAUL" --broker 123

  # Create with description
  xbe do work-order-service-codes create --code "SPREAD" --broker 123 --description "Spreading service"

  # Get JSON output
  xbe do work-order-service-codes create --code "TEST" --broker 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoWorkOrderServiceCodesCreate,
	}
	initDoWorkOrderServiceCodesCreateFlags(cmd)
	return cmd
}

func init() {
	doWorkOrderServiceCodesCmd.AddCommand(newDoWorkOrderServiceCodesCreateCmd())
}

func initDoWorkOrderServiceCodesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("code", "", "Service code value (required)")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoWorkOrderServiceCodesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoWorkOrderServiceCodesCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Code) == "" {
		err := fmt.Errorf("--code is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Broker) == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"code": opts.Code,
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "work-order-service-codes",
			"attributes": attributes,
			"relationships": map[string]any{
				"broker": map[string]any{
					"data": map[string]any{
						"type": "brokers",
						"id":   opts.Broker,
					},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/work-order-service-codes", jsonBody)
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

	row := buildWorkOrderServiceCodeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created work order service code %s (%s)\n", row.ID, row.Code)
	return nil
}

func parseDoWorkOrderServiceCodesCreateOptions(cmd *cobra.Command) (doWorkOrderServiceCodesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	code, _ := cmd.Flags().GetString("code")
	description, _ := cmd.Flags().GetString("description")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doWorkOrderServiceCodesCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		Code:        code,
		Description: description,
		Broker:      broker,
	}, nil
}

func buildWorkOrderServiceCodeRowFromSingle(resp jsonAPISingleResponse) workOrderServiceCodeRow {
	attrs := resp.Data.Attributes

	row := workOrderServiceCodeRow{
		ID:          resp.Data.ID,
		Code:        stringAttr(attrs, "code"),
		Description: stringAttr(attrs, "description"),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}

	return row
}
