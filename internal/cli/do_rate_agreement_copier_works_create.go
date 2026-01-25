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

type doRateAgreementCopierWorksCreateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	RateAgreementTemplate  string
	TargetOrganizationType string
	TargetOrganizationID   string
	Note                   string
}

func newDoRateAgreementCopierWorksCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a rate agreement copier work",
		Long: `Create a rate agreement copier work to copy a rate agreement to a target organization.

Required flags:
  --rate-agreement-template   Rate agreement template ID (required)
  --target-organization-type  Target organization type (Customer, Trucker) (required)
  --target-organization-id    Target organization ID (required)

Optional flags:
  --note                      Add a note to the copier work`,
		Example: `  # Copy a rate agreement to a customer
  xbe do rate-agreement-copier-works create \
    --rate-agreement-template 123 \
    --target-organization-type Customer \
    --target-organization-id 456 \
    --note "Copy template to new customer"`,
		Args: cobra.NoArgs,
		RunE: runDoRateAgreementCopierWorksCreate,
	}
	initDoRateAgreementCopierWorksCreateFlags(cmd)
	return cmd
}

func init() {
	doRateAgreementCopierWorksCmd.AddCommand(newDoRateAgreementCopierWorksCreateCmd())
}

func initDoRateAgreementCopierWorksCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("rate-agreement-template", "", "Rate agreement template ID (required)")
	cmd.Flags().String("target-organization-type", "", "Target organization type (Customer, Trucker) (required)")
	cmd.Flags().String("target-organization-id", "", "Target organization ID (required)")
	cmd.Flags().String("note", "", "Optional note")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoRateAgreementCopierWorksCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoRateAgreementCopierWorksCreateOptions(cmd)
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

	if opts.RateAgreementTemplate == "" {
		err := fmt.Errorf("--rate-agreement-template is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.TargetOrganizationType == "" {
		err := fmt.Errorf("--target-organization-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.TargetOrganizationID == "" {
		err := fmt.Errorf("--target-organization-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.Note != "" {
		attributes["note"] = opts.Note
	}

	targetType := normalizeTargetOrganizationTypeForRelationship(opts.TargetOrganizationType)

	relationships := map[string]any{
		"rate-agreement-template": map[string]any{
			"data": map[string]any{
				"type": "rate-agreements",
				"id":   opts.RateAgreementTemplate,
			},
		},
		"target-organization": map[string]any{
			"data": map[string]any{
				"type": targetType,
				"id":   opts.TargetOrganizationID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "rate-agreement-copier-works",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/rate-agreement-copier-works", jsonBody)
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

	row := buildRateAgreementCopierWorkRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created rate agreement copier work %s\n", row.ID)
	return nil
}

func parseDoRateAgreementCopierWorksCreateOptions(cmd *cobra.Command) (doRateAgreementCopierWorksCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	rateAgreementTemplate, _ := cmd.Flags().GetString("rate-agreement-template")
	targetOrganizationType, _ := cmd.Flags().GetString("target-organization-type")
	targetOrganizationID, _ := cmd.Flags().GetString("target-organization-id")
	note, _ := cmd.Flags().GetString("note")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRateAgreementCopierWorksCreateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		RateAgreementTemplate:  rateAgreementTemplate,
		TargetOrganizationType: targetOrganizationType,
		TargetOrganizationID:   targetOrganizationID,
		Note:                   note,
	}, nil
}

func buildRateAgreementCopierWorkRowFromSingle(resp jsonAPISingleResponse) rateAgreementCopierWorkRow {
	resource := resp.Data
	row := rateAgreementCopierWorkRow{
		ID:          resource.ID,
		Note:        stringAttr(resource.Attributes, "note"),
		ScheduledAt: formatDateTime(stringAttr(resource.Attributes, "scheduled-at")),
		ProcessedAt: formatDateTime(stringAttr(resource.Attributes, "processed-at")),
	}

	if rel, ok := resource.Relationships["rate-agreement-template"]; ok && rel.Data != nil {
		row.RateAgreementTemplateID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["target-organization"]; ok && rel.Data != nil {
		row.TargetOrganizationType = rel.Data.Type
		row.TargetOrganizationID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	return row
}
