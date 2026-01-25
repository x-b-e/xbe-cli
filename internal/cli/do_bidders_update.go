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

type doBiddersUpdateOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	ID              string
	Name            string
	IsSelfForBroker bool
}

func newDoBiddersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing bidder",
		Long: `Update an existing bidder.

Provide the bidder ID as an argument, then use flags to specify which
fields to update. Only specified fields will be modified.

Updatable fields:
  --name                Bidder name
  --is-self-for-broker  Mark as the broker's self bidder (true/false)`,
		Example: `  # Update bidder name
  xbe do bidders update 123 --name "Acme Logistics West"

  # Update self bidder status
  xbe do bidders update 123 --is-self-for-broker true

  # Update multiple fields
  xbe do bidders update 123 --name "Acme Logistics" --is-self-for-broker false

  # Get JSON output
  xbe do bidders update 123 --name "Updated" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoBiddersUpdate,
	}
	initDoBiddersUpdateFlags(cmd)
	return cmd
}

func init() {
	doBiddersCmd.AddCommand(newDoBiddersUpdateCmd())
}

func initDoBiddersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Bidder name")
	cmd.Flags().Bool("is-self-for-broker", false, "Mark as the broker's self bidder")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBiddersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoBiddersUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
	}
	if cmd.Flags().Changed("is-self-for-broker") {
		attributes["is-self-for-broker"] = opts.IsSelfForBroker
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one of --name, --is-self-for-broker")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "bidders",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/bidders/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated bidder %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoBiddersUpdateOptions(cmd *cobra.Command, args []string) (doBiddersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	isSelfForBroker, _ := cmd.Flags().GetBool("is-self-for-broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBiddersUpdateOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		ID:              args[0],
		Name:            name,
		IsSelfForBroker: isSelfForBroker,
	}, nil
}
