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

type rateAgreementCopierWorksShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type rateAgreementCopierWorkDetails struct {
	ID                      string `json:"id"`
	RateAgreementTemplateID string `json:"rate_agreement_template_id,omitempty"`
	TargetOrganizationType  string `json:"target_organization_type,omitempty"`
	TargetOrganizationID    string `json:"target_organization_id,omitempty"`
	BrokerID                string `json:"broker_id,omitempty"`
	CreatedByID             string `json:"created_by_id,omitempty"`
	Note                    string `json:"note,omitempty"`
	ScheduledAt             string `json:"scheduled_at,omitempty"`
	ProcessedAt             string `json:"processed_at,omitempty"`
	WorkResults             any    `json:"work_results,omitempty"`
	WorkErrors              any    `json:"work_errors,omitempty"`
	CreatedAt               string `json:"created_at,omitempty"`
	UpdatedAt               string `json:"updated_at,omitempty"`
}

func newRateAgreementCopierWorksShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show rate agreement copier work details",
		Long: `Show the full details of a rate agreement copier work item.

Output Fields:
  ID
  Rate Agreement Template ID
  Target Organization (type and ID)
  Broker (ID)
  Created By (user ID)
  Note
  Scheduled At
  Processed At
  Work Results
  Work Errors
  Created At
  Updated At

Arguments:
  <id>    The copier work ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show copier work details
  xbe view rate-agreement-copier-works show 123

  # Get JSON output
  xbe view rate-agreement-copier-works show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runRateAgreementCopierWorksShow,
	}
	initRateAgreementCopierWorksShowFlags(cmd)
	return cmd
}

func init() {
	rateAgreementCopierWorksCmd.AddCommand(newRateAgreementCopierWorksShowCmd())
}

func initRateAgreementCopierWorksShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRateAgreementCopierWorksShow(cmd *cobra.Command, args []string) error {
	opts, err := parseRateAgreementCopierWorksShowOptions(cmd)
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
		return fmt.Errorf("rate agreement copier work id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/rate-agreement-copier-works/"+id, nil)
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

	details := buildRateAgreementCopierWorkDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderRateAgreementCopierWorkDetails(cmd, details)
}

func parseRateAgreementCopierWorksShowOptions(cmd *cobra.Command) (rateAgreementCopierWorksShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return rateAgreementCopierWorksShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildRateAgreementCopierWorkDetails(resp jsonAPISingleResponse) rateAgreementCopierWorkDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := rateAgreementCopierWorkDetails{
		ID:          resource.ID,
		Note:        stringAttr(attrs, "note"),
		ScheduledAt: formatDateTime(stringAttr(attrs, "scheduled-at")),
		ProcessedAt: formatDateTime(stringAttr(attrs, "processed-at")),
		WorkResults: anyAttr(attrs, "work-results"),
		WorkErrors:  anyAttr(attrs, "work-errors"),
		CreatedAt:   formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:   formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["rate-agreement-template"]; ok && rel.Data != nil {
		details.RateAgreementTemplateID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["target-organization"]; ok && rel.Data != nil {
		details.TargetOrganizationType = rel.Data.Type
		details.TargetOrganizationID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}

	return details
}

func renderRateAgreementCopierWorkDetails(cmd *cobra.Command, details rateAgreementCopierWorkDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.RateAgreementTemplateID != "" {
		fmt.Fprintf(out, "Rate Agreement Template ID: %s\n", details.RateAgreementTemplateID)
	}
	if details.TargetOrganizationType != "" && details.TargetOrganizationID != "" {
		fmt.Fprintf(out, "Target Organization: %s/%s\n", details.TargetOrganizationType, details.TargetOrganizationID)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
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
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}
	if details.WorkResults != nil {
		fmt.Fprintln(out, "Work Results:")
		if err := writeJSON(out, details.WorkResults); err != nil {
			return err
		}
	}
	if details.WorkErrors != nil {
		fmt.Fprintln(out, "Work Errors:")
		if err := writeJSON(out, details.WorkErrors); err != nil {
			return err
		}
	}

	return nil
}
