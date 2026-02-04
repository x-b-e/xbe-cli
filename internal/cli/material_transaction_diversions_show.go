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

type materialTransactionDiversionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type materialTransactionDiversionDetails struct {
	ID                              string `json:"id"`
	MaterialTransactionID           string `json:"material_transaction_id,omitempty"`
	MaterialTransactionTicketNumber string `json:"material_transaction_ticket_number,omitempty"`
	NewJobSiteID                    string `json:"new_job_site_id,omitempty"`
	NewJobSiteName                  string `json:"new_job_site_name,omitempty"`
	NewDeliveryDate                 string `json:"new_delivery_date,omitempty"`
	DivertedTonsExplicit            string `json:"diverted_tons_explicit,omitempty"`
	DivertedTons                    string `json:"diverted_tons,omitempty"`
	DriverInstructions              string `json:"driver_instructions,omitempty"`
	BrokerID                        string `json:"broker_id,omitempty"`
	BrokerName                      string `json:"broker_name,omitempty"`
	CreatedByID                     string `json:"created_by_id,omitempty"`
	CreatedByName                   string `json:"created_by_name,omitempty"`
}

func newMaterialTransactionDiversionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show material transaction diversion details",
		Long: `Show the full details of a material transaction diversion.

Arguments:
  <id>  The diversion ID (required).`,
		Example: `  # Show a diversion
  xbe view material-transaction-diversions show 123

  # Output as JSON
  xbe view material-transaction-diversions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaterialTransactionDiversionsShow,
	}
	initMaterialTransactionDiversionsShowFlags(cmd)
	return cmd
}

func init() {
	materialTransactionDiversionsCmd.AddCommand(newMaterialTransactionDiversionsShowCmd())
}

func initMaterialTransactionDiversionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTransactionDiversionsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseMaterialTransactionDiversionsShowOptions(cmd)
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
		return fmt.Errorf("material transaction diversion id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-transaction-diversions]", "material-transaction,new-job-site,new-delivery-date,diverted-tons-explicit,diverted-tons,driver-instructions,created-by,broker")
	query.Set("include", "material-transaction,new-job-site,created-by,broker")
	query.Set("fields[material-transactions]", "ticket-number")
	query.Set("fields[job-sites]", "name")
	query.Set("fields[users]", "name")
	query.Set("fields[brokers]", "company-name")

	body, _, err := client.Get(cmd.Context(), "/v1/material-transaction-diversions/"+id, query)
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

	details := buildMaterialTransactionDiversionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaterialTransactionDiversionDetails(cmd, details)
}

func parseMaterialTransactionDiversionsShowOptions(cmd *cobra.Command) (materialTransactionDiversionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialTransactionDiversionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaterialTransactionDiversionDetails(resp jsonAPISingleResponse) materialTransactionDiversionDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := materialTransactionDiversionDetails{
		ID:                   resp.Data.ID,
		NewDeliveryDate:      formatDate(stringAttr(attrs, "new-delivery-date")),
		DivertedTonsExplicit: stringAttr(attrs, "diverted-tons-explicit"),
		DivertedTons:         stringAttr(attrs, "diverted-tons"),
		DriverInstructions:   stringAttr(attrs, "driver-instructions"),
	}

	if rel, ok := resp.Data.Relationships["material-transaction"]; ok && rel.Data != nil {
		details.MaterialTransactionID = rel.Data.ID
		if mtxn, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialTransactionTicketNumber = stringAttr(mtxn.Attributes, "ticket-number")
		}
	}

	if rel, ok := resp.Data.Relationships["new-job-site"]; ok && rel.Data != nil {
		details.NewJobSiteID = rel.Data.ID
		if site, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.NewJobSiteName = stringAttr(site.Attributes, "name")
		}
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BrokerName = firstNonEmpty(
				stringAttr(broker.Attributes, "company-name"),
				stringAttr(broker.Attributes, "name"),
			)
		}
	}

	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.CreatedByName = stringAttr(user.Attributes, "name")
		}
	}

	return details
}

func renderMaterialTransactionDiversionDetails(cmd *cobra.Command, details materialTransactionDiversionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.MaterialTransactionID != "" || details.MaterialTransactionTicketNumber != "" {
		fmt.Fprintf(out, "Material Transaction: %s\n", formatRelated(details.MaterialTransactionTicketNumber, details.MaterialTransactionID))
	}
	if details.NewJobSiteID != "" || details.NewJobSiteName != "" {
		fmt.Fprintf(out, "New Job Site: %s\n", formatRelated(details.NewJobSiteName, details.NewJobSiteID))
	}
	if details.NewDeliveryDate != "" {
		fmt.Fprintf(out, "New Delivery Date: %s\n", details.NewDeliveryDate)
	}
	if details.DivertedTons != "" {
		fmt.Fprintf(out, "Diverted Tons: %s\n", details.DivertedTons)
	}
	if details.DivertedTonsExplicit != "" {
		fmt.Fprintf(out, "Diverted Tons (Explicit): %s\n", details.DivertedTonsExplicit)
	}
	if details.DriverInstructions != "" {
		fmt.Fprintf(out, "Driver Instructions: %s\n", details.DriverInstructions)
	}
	if details.BrokerID != "" || details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", formatRelated(details.BrokerName, details.BrokerID))
	}
	if details.CreatedByID != "" || details.CreatedByName != "" {
		fmt.Fprintf(out, "Created By: %s\n", formatRelated(details.CreatedByName, details.CreatedByID))
	}

	return nil
}
