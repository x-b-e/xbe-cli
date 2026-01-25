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

type doLineupDispatchesUpdateOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	ID                       string
	Comment                  string
	AutoOfferCustomerTenders bool
	AutoOfferTruckerTenders  bool
	AutoAcceptTruckerTenders bool
}

func newDoLineupDispatchesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a lineup dispatch",
		Long: `Update a lineup dispatch.

Optional flags:
  --comment                      Dispatch comment
  --auto-offer-customer-tenders  Auto-offer customer tenders
  --auto-offer-trucker-tenders   Auto-offer trucker tenders
  --auto-accept-trucker-tenders  Auto-accept trucker tenders`,
		Example: `  # Update a dispatch comment
  xbe do lineup-dispatches update 123 --comment "Updated note"

  # Toggle tender settings
  xbe do lineup-dispatches update 123 \
    --auto-offer-customer-tenders=false \
    --auto-offer-trucker-tenders=false \
    --auto-accept-trucker-tenders=true`,
		Args: cobra.ExactArgs(1),
		RunE: runDoLineupDispatchesUpdate,
	}
	initDoLineupDispatchesUpdateFlags(cmd)
	return cmd
}

func init() {
	doLineupDispatchesCmd.AddCommand(newDoLineupDispatchesUpdateCmd())
}

func initDoLineupDispatchesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("comment", "", "Dispatch comment")
	cmd.Flags().Bool("auto-offer-customer-tenders", false, "Auto-offer customer tenders")
	cmd.Flags().Bool("auto-offer-trucker-tenders", false, "Auto-offer trucker tenders")
	cmd.Flags().Bool("auto-accept-trucker-tenders", false, "Auto-accept trucker tenders")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLineupDispatchesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoLineupDispatchesUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("comment") {
		attributes["comment"] = opts.Comment
	}
	if cmd.Flags().Changed("auto-offer-customer-tenders") {
		attributes["auto-offer-customer-tenders"] = opts.AutoOfferCustomerTenders
	}
	if cmd.Flags().Changed("auto-offer-trucker-tenders") {
		attributes["auto-offer-trucker-tenders"] = opts.AutoOfferTruckerTenders
	}
	if cmd.Flags().Changed("auto-accept-trucker-tenders") {
		attributes["auto-accept-trucker-tenders"] = opts.AutoAcceptTruckerTenders
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "lineup-dispatches",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
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

	body, _, err := client.Patch(cmd.Context(), "/v1/lineup-dispatches/"+opts.ID, jsonBody)
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

	row := lineupDispatchRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated lineup dispatch %s\n", row.ID)
	return nil
}

func parseDoLineupDispatchesUpdateOptions(cmd *cobra.Command, args []string) (doLineupDispatchesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	comment, _ := cmd.Flags().GetString("comment")
	autoOfferCustomerTenders, _ := cmd.Flags().GetBool("auto-offer-customer-tenders")
	autoOfferTruckerTenders, _ := cmd.Flags().GetBool("auto-offer-trucker-tenders")
	autoAcceptTruckerTenders, _ := cmd.Flags().GetBool("auto-accept-trucker-tenders")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLineupDispatchesUpdateOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		ID:                       args[0],
		Comment:                  comment,
		AutoOfferCustomerTenders: autoOfferCustomerTenders,
		AutoOfferTruckerTenders:  autoOfferTruckerTenders,
		AutoAcceptTruckerTenders: autoAcceptTruckerTenders,
	}, nil
}
