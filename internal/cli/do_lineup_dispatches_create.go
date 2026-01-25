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

type doLineupDispatchesCreateOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	Lineup                   string
	Comment                  string
	AutoOfferCustomerTenders bool
	AutoOfferTruckerTenders  bool
	AutoAcceptTruckerTenders bool
}

func newDoLineupDispatchesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a lineup dispatch",
		Long: `Create a lineup dispatch.

Required flags:
  --lineup  Lineup ID (required)

Optional flags:
  --comment                      Dispatch comment
  --auto-offer-customer-tenders  Auto-offer customer tenders
  --auto-offer-trucker-tenders   Auto-offer trucker tenders
  --auto-accept-trucker-tenders  Auto-accept trucker tenders`,
		Example: `  # Create a lineup dispatch for a lineup
  xbe do lineup-dispatches create --lineup 123

  # Create with custom tender settings
  xbe do lineup-dispatches create --lineup 123 \
    --auto-offer-customer-tenders=false \
    --auto-offer-trucker-tenders=false \
    --auto-accept-trucker-tenders=true

  # JSON output
  xbe do lineup-dispatches create --lineup 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoLineupDispatchesCreate,
	}
	initDoLineupDispatchesCreateFlags(cmd)
	return cmd
}

func init() {
	doLineupDispatchesCmd.AddCommand(newDoLineupDispatchesCreateCmd())
}

func initDoLineupDispatchesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("lineup", "", "Lineup ID (required)")
	cmd.Flags().String("comment", "", "Dispatch comment")
	cmd.Flags().Bool("auto-offer-customer-tenders", false, "Auto-offer customer tenders")
	cmd.Flags().Bool("auto-offer-trucker-tenders", false, "Auto-offer trucker tenders")
	cmd.Flags().Bool("auto-accept-trucker-tenders", false, "Auto-accept trucker tenders")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLineupDispatchesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoLineupDispatchesCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Lineup) == "" {
		err := fmt.Errorf("--lineup is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
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

	relationships := map[string]any{
		"lineup": map[string]any{
			"data": map[string]any{
				"type": "lineups",
				"id":   opts.Lineup,
			},
		},
	}

	data := map[string]any{
		"type":          "lineup-dispatches",
		"relationships": relationships,
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

	body, _, err := client.Post(cmd.Context(), "/v1/lineup-dispatches", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created lineup dispatch %s\n", row.ID)
	return nil
}

func parseDoLineupDispatchesCreateOptions(cmd *cobra.Command) (doLineupDispatchesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	lineup, _ := cmd.Flags().GetString("lineup")
	comment, _ := cmd.Flags().GetString("comment")
	autoOfferCustomerTenders, _ := cmd.Flags().GetBool("auto-offer-customer-tenders")
	autoOfferTruckerTenders, _ := cmd.Flags().GetBool("auto-offer-trucker-tenders")
	autoAcceptTruckerTenders, _ := cmd.Flags().GetBool("auto-accept-trucker-tenders")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLineupDispatchesCreateOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		Lineup:                   lineup,
		Comment:                  comment,
		AutoOfferCustomerTenders: autoOfferCustomerTenders,
		AutoOfferTruckerTenders:  autoOfferTruckerTenders,
		AutoAcceptTruckerTenders: autoAcceptTruckerTenders,
	}, nil
}

func lineupDispatchRowFromSingle(resp jsonAPISingleResponse) lineupDispatchRow {
	attrs := resp.Data.Attributes
	row := lineupDispatchRow{
		ID:               resp.Data.ID,
		IsFulfilled:      boolAttr(attrs, "is-fulfilled"),
		IsFulfilling:     boolAttr(attrs, "is-fulfilling"),
		FulfillmentCount: intAttr(attrs, "fulfillment-count"),
	}

	if rel, ok := resp.Data.Relationships["lineup"]; ok && rel.Data != nil {
		row.LineupID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	return row
}
