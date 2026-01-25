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

type doBiddersCreateOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	Name            string
	Broker          string
	IsSelfForBroker bool
}

func newDoBiddersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new bidder",
		Long: `Create a new bidder.

Required flags:
  --name                Bidder name (required)
  --broker              Broker ID (required)
  --is-self-for-broker  Mark as the broker's self bidder (true/false)

Optional flags:
  --json  Output JSON`,
		Example: `  # Create a bidder
  xbe do bidders create --name "Acme Logistics" --broker 123 --is-self-for-broker false

  # Create the broker's self bidder
  xbe do bidders create --name "Acme Logistics" --broker 123 --is-self-for-broker true

  # Get JSON output
  xbe do bidders create --name "Acme Logistics" --broker 123 --is-self-for-broker false --json`,
		Args: cobra.NoArgs,
		RunE: runDoBiddersCreate,
	}
	initDoBiddersCreateFlags(cmd)
	return cmd
}

func init() {
	doBiddersCmd.AddCommand(newDoBiddersCreateCmd())
}

func initDoBiddersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Bidder name (required)")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().Bool("is-self-for-broker", false, "Mark as the broker's self bidder")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBiddersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoBiddersCreateOptions(cmd)
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

	if opts.Name == "" {
		err := fmt.Errorf("--name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.Broker == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !cmd.Flags().Changed("is-self-for-broker") {
		err := fmt.Errorf("--is-self-for-broker is required (true/false)")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"name":               opts.Name,
		"is-self-for-broker": opts.IsSelfForBroker,
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "bidders",
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

	body, _, err := client.Post(cmd.Context(), "/v1/bidders", jsonBody)
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

	row := buildBidderRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created bidder %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoBiddersCreateOptions(cmd *cobra.Command) (doBiddersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	broker, _ := cmd.Flags().GetString("broker")
	isSelfForBroker, _ := cmd.Flags().GetBool("is-self-for-broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBiddersCreateOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		Name:            name,
		Broker:          broker,
		IsSelfForBroker: isSelfForBroker,
	}, nil
}

func buildBidderRowFromSingle(resp jsonAPISingleResponse) bidderRow {
	attrs := resp.Data.Attributes

	row := bidderRow{
		ID:              resp.Data.ID,
		Name:            stringAttr(attrs, "name"),
		IsSelfForBroker: boolAttr(attrs, "is-self-for-broker"),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}

	return row
}
