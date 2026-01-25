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

type doRateAgreementCopiersCreateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	TemplateRateAgreement  string
	TargetOrganizationType string
	TargetOrganizationID   string
}

type rateAgreementCopierRow struct {
	ID                      string `json:"id"`
	TemplateRateAgreementID string `json:"template_rate_agreement_id,omitempty"`
	TargetOrganizationType  string `json:"target_organization_type,omitempty"`
	TargetOrganizationID    string `json:"target_organization_id,omitempty"`
	TargetRateAgreementID   string `json:"target_rate_agreement_id,omitempty"`
}

func newDoRateAgreementCopiersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Copy a rate agreement to a target organization",
		Long: `Copy a rate agreement to a target organization.

Required flags:
  --template-rate-agreement     Template rate agreement ID (required)
  --target-organization-type    Target organization type (customers, truckers) (required)
  --target-organization-id      Target organization ID (required)

The template rate agreement must match the target organization type
(customer or trucker).`,
		Example: `  # Copy a template rate agreement to a customer
  xbe do rate-agreement-copiers create \
    --template-rate-agreement 123 \
    --target-organization-type customers \
    --target-organization-id 456

  # Copy a template rate agreement to a trucker
  xbe do rate-agreement-copiers create \
    --template-rate-agreement 123 \
    --target-organization-type truckers \
    --target-organization-id 789

  # Output as JSON
  xbe do rate-agreement-copiers create \
    --template-rate-agreement 123 \
    --target-organization-type customers \
    --target-organization-id 456 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoRateAgreementCopiersCreate,
	}
	initDoRateAgreementCopiersCreateFlags(cmd)
	return cmd
}

func init() {
	doRateAgreementCopiersCmd.AddCommand(newDoRateAgreementCopiersCreateCmd())
}

func initDoRateAgreementCopiersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("template-rate-agreement", "", "Template rate agreement ID (required)")
	cmd.Flags().String("target-organization-type", "", "Target organization type (customers, truckers) (required)")
	cmd.Flags().String("target-organization-id", "", "Target organization ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoRateAgreementCopiersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoRateAgreementCopiersCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run xbe auth login first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	opts.TemplateRateAgreement = strings.TrimSpace(opts.TemplateRateAgreement)
	if opts.TemplateRateAgreement == "" {
		err := fmt.Errorf("--template-rate-agreement is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	opts.TargetOrganizationType = strings.TrimSpace(opts.TargetOrganizationType)
	if opts.TargetOrganizationType == "" {
		err := fmt.Errorf("--target-organization-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	opts.TargetOrganizationID = strings.TrimSpace(opts.TargetOrganizationID)
	if opts.TargetOrganizationID == "" {
		err := fmt.Errorf("--target-organization-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"template-rate-agreement": map[string]any{
			"data": map[string]any{
				"type": "rate-agreements",
				"id":   opts.TemplateRateAgreement,
			},
		},
		"target-organization": map[string]any{
			"data": map[string]any{
				"type": opts.TargetOrganizationType,
				"id":   opts.TargetOrganizationID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "rate-agreement-copiers",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/rate-agreement-copiers", jsonBody)
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

	row := buildRateAgreementCopierRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created rate agreement copier %s\n", row.ID)
	return nil
}

func buildRateAgreementCopierRowFromSingle(resp jsonAPISingleResponse) rateAgreementCopierRow {
	row := rateAgreementCopierRow{
		ID: resp.Data.ID,
	}

	if rel, ok := resp.Data.Relationships["template-rate-agreement"]; ok && rel.Data != nil {
		row.TemplateRateAgreementID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["target-organization"]; ok && rel.Data != nil {
		row.TargetOrganizationType = rel.Data.Type
		row.TargetOrganizationID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["target-rate-agreement"]; ok && rel.Data != nil {
		row.TargetRateAgreementID = rel.Data.ID
	}

	return row
}

func parseDoRateAgreementCopiersCreateOptions(cmd *cobra.Command) (doRateAgreementCopiersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	templateRateAgreement, _ := cmd.Flags().GetString("template-rate-agreement")
	targetOrganizationType, _ := cmd.Flags().GetString("target-organization-type")
	targetOrganizationID, _ := cmd.Flags().GetString("target-organization-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRateAgreementCopiersCreateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		TemplateRateAgreement:  templateRateAgreement,
		TargetOrganizationType: targetOrganizationType,
		TargetOrganizationID:   targetOrganizationID,
	}, nil
}
