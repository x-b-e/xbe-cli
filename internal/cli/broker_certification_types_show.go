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

type brokerCertificationTypesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type brokerCertificationTypeDetails struct {
	ID                  string `json:"id"`
	BrokerID            string `json:"broker_id,omitempty"`
	CertificationTypeID string `json:"certification_type_id,omitempty"`
}

func newBrokerCertificationTypesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show broker certification type details",
		Long: `Show the full details of a broker certification type.

Output Fields:
  ID             Broker certification type identifier
  BROKER ID      Broker ID
  CERT TYPE ID   Certification type ID

Arguments:
  <id>  The broker certification type ID (required). Use the list command to find IDs.`,
		Example: `  # Show a broker certification type
  xbe view broker-certification-types show 123

  # Show as JSON
  xbe view broker-certification-types show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runBrokerCertificationTypesShow,
	}
	initBrokerCertificationTypesShowFlags(cmd)
	return cmd
}

func init() {
	brokerCertificationTypesCmd.AddCommand(newBrokerCertificationTypesShowCmd())
}

func initBrokerCertificationTypesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBrokerCertificationTypesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseBrokerCertificationTypesShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("broker certification type id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[broker-certification-types]", "broker,certification-type")

	body, _, err := client.Get(cmd.Context(), "/v1/broker-certification-types/"+id, query)
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

	details := buildBrokerCertificationTypeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderBrokerCertificationTypeDetails(cmd, details)
}

func parseBrokerCertificationTypesShowOptions(cmd *cobra.Command) (brokerCertificationTypesShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return brokerCertificationTypesShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return brokerCertificationTypesShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return brokerCertificationTypesShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return brokerCertificationTypesShowOptions{}, err
	}

	return brokerCertificationTypesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildBrokerCertificationTypeDetails(resp jsonAPISingleResponse) brokerCertificationTypeDetails {
	details := brokerCertificationTypeDetails{
		ID: resp.Data.ID,
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["certification-type"]; ok && rel.Data != nil {
		details.CertificationTypeID = rel.Data.ID
	}

	return details
}

func renderBrokerCertificationTypeDetails(cmd *cobra.Command, details brokerCertificationTypeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.CertificationTypeID != "" {
		fmt.Fprintf(out, "Certification Type ID: %s\n", details.CertificationTypeID)
	}

	return nil
}
