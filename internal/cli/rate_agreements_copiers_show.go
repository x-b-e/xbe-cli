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

type rateAgreementsCopiersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type rateAgreementsCopierDetails struct {
	ID                         string         `json:"id"`
	RateAgreementTemplateID    string         `json:"rate_agreement_template_id,omitempty"`
	BrokerID                   string         `json:"broker_id,omitempty"`
	CreatedByID                string         `json:"created_by_id,omitempty"`
	TargetCustomerIDs          []string       `json:"target_customer_ids,omitempty"`
	TargetTruckerIDs           []string       `json:"target_trucker_ids,omitempty"`
	RateAgreementCopierWorkIDs []string       `json:"rate_agreement_copier_work_ids,omitempty"`
	Note                       string         `json:"note,omitempty"`
	ScheduledAt                string         `json:"scheduled_at,omitempty"`
	ProcessedAt                string         `json:"processed_at,omitempty"`
	CopiersResults             map[string]any `json:"copiers_results,omitempty"`
	CopiersErrors              map[string]any `json:"copiers_errors,omitempty"`
}

func newRateAgreementsCopiersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show rate agreements copier details",
		Long: `Show the full details of a rate agreements copier.

Output Fields:
  ID
  Rate Agreement Template ID
  Broker ID
  Created By ID
  Target Customer IDs
  Target Trucker IDs
  Rate Agreement Copier Work IDs
  Note
  Scheduled At
  Processed At
  Copiers Results
  Copiers Errors

Arguments:
  <id>    The rate agreements copier ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a rate agreements copier
  xbe view rate-agreements-copiers show 123

  # Output as JSON
  xbe view rate-agreements-copiers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runRateAgreementsCopiersShow,
	}
	initRateAgreementsCopiersShowFlags(cmd)
	return cmd
}

func init() {
	rateAgreementsCopiersCmd.AddCommand(newRateAgreementsCopiersShowCmd())
}

func initRateAgreementsCopiersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRateAgreementsCopiersShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseRateAgreementsCopiersShowOptions(cmd)
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
		return fmt.Errorf("rate agreements copier id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[rate-agreements-copiers]", "note,scheduled-at,processed-at,copiers-results,copiers-errors,rate-agreement-template,broker,created-by,target-customers,target-truckers,rate-agreement-copier-works")

	body, _, err := client.Get(cmd.Context(), "/v1/rate-agreements-copiers/"+id, query)
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

	details := buildRateAgreementsCopierDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderRateAgreementsCopierDetails(cmd, details)
}

func parseRateAgreementsCopiersShowOptions(cmd *cobra.Command) (rateAgreementsCopiersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return rateAgreementsCopiersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildRateAgreementsCopierDetails(resp jsonAPISingleResponse) rateAgreementsCopierDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := rateAgreementsCopierDetails{
		ID:             resource.ID,
		Note:           stringAttr(attrs, "note"),
		ScheduledAt:    formatDateTime(stringAttr(attrs, "scheduled-at")),
		ProcessedAt:    formatDateTime(stringAttr(attrs, "processed-at")),
		CopiersResults: mapAttr(attrs, "copiers-results"),
		CopiersErrors:  mapAttr(attrs, "copiers-errors"),
	}

	if rel, ok := resource.Relationships["rate-agreement-template"]; ok && rel.Data != nil {
		details.RateAgreementTemplateID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["target-customers"]; ok {
		details.TargetCustomerIDs = relationshipIDsToStrings(rel)
	}
	if rel, ok := resource.Relationships["target-truckers"]; ok {
		details.TargetTruckerIDs = relationshipIDsToStrings(rel)
	}
	if rel, ok := resource.Relationships["rate-agreement-copier-works"]; ok {
		details.RateAgreementCopierWorkIDs = relationshipIDsToStrings(rel)
	}

	return details
}

func renderRateAgreementsCopierDetails(cmd *cobra.Command, details rateAgreementsCopierDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.RateAgreementTemplateID != "" {
		fmt.Fprintf(out, "Rate Agreement Template ID: %s\n", details.RateAgreementTemplateID)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By ID: %s\n", details.CreatedByID)
	}
	if len(details.TargetCustomerIDs) > 0 {
		fmt.Fprintf(out, "Target Customer IDs: %s\n", strings.Join(details.TargetCustomerIDs, ", "))
	}
	if len(details.TargetTruckerIDs) > 0 {
		fmt.Fprintf(out, "Target Trucker IDs: %s\n", strings.Join(details.TargetTruckerIDs, ", "))
	}
	if len(details.RateAgreementCopierWorkIDs) > 0 {
		fmt.Fprintf(out, "Rate Agreement Copier Work IDs: %s\n", strings.Join(details.RateAgreementCopierWorkIDs, ", "))
	}
	if details.Note != "" {
		fmt.Fprintf(out, "Note: %s\n", details.Note)
	}
	if details.ScheduledAt != "" {
		fmt.Fprintf(out, "Scheduled At: %s\n", details.ScheduledAt)
	}
	if details.ProcessedAt != "" {
		fmt.Fprintf(out, "Processed At: %s\n", details.ProcessedAt)
	}
	if len(details.CopiersResults) > 0 {
		fmt.Fprintf(out, "Copiers Results: %s\n", formatMap(details.CopiersResults))
	}
	if len(details.CopiersErrors) > 0 {
		fmt.Fprintf(out, "Copiers Errors: %s\n", formatMap(details.CopiersErrors))
	}

	return nil
}
