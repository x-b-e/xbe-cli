package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type truckerReferralCodesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type truckerReferralCodeDetails struct {
	ID       string `json:"id"`
	Code     string `json:"code,omitempty"`
	Value    string `json:"value,omitempty"`
	BrokerID string `json:"broker_id,omitempty"`
}

func newTruckerReferralCodesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show trucker referral code details",
		Long: `Show the full details of a trucker referral code.

Output Fields:
  ID         Trucker referral code identifier
  CODE       Referral code (normalized to uppercase, no spaces)
  VALUE      Referral value
  BROKER ID  Broker ID

Arguments:
  <id>  The trucker referral code ID (required). Use the list command to find IDs.`,
		Example: `  # Show a trucker referral code
  xbe view trucker-referral-codes show 123

  # Show as JSON
  xbe view trucker-referral-codes show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTruckerReferralCodesShow,
	}
	initTruckerReferralCodesShowFlags(cmd)
	return cmd
}

func init() {
	truckerReferralCodesCmd.AddCommand(newTruckerReferralCodesShowCmd())
}

func initTruckerReferralCodesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTruckerReferralCodesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseTruckerReferralCodesShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("trucker referral code id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[trucker-referral-codes]", "code,value,broker")

	body, _, err := client.Get(cmd.Context(), "/v1/trucker-referral-codes/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildTruckerReferralCodeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTruckerReferralCodeDetails(cmd, details)
}

func parseTruckerReferralCodesShowOptions(cmd *cobra.Command) (truckerReferralCodesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return truckerReferralCodesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTruckerReferralCodeDetails(resp jsonAPISingleResponse) truckerReferralCodeDetails {
	details := truckerReferralCodeDetails{
		ID:    resp.Data.ID,
		Code:  stringAttr(resp.Data.Attributes, "code"),
		Value: stringAttr(resp.Data.Attributes, "value"),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}

	return details
}

func renderTruckerReferralCodeDetails(cmd *cobra.Command, details truckerReferralCodeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Code != "" {
		fmt.Fprintf(out, "Code: %s\n", details.Code)
	}
	if details.Value != "" {
		fmt.Fprintf(out, "Value: %s\n", details.Value)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}

	return nil
}
