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

type doProfitImprovementsCreateOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	Title                     string
	Description               string
	Status                    string
	AmountEstimated           float64
	ImpactFrequencyEstimated  string
	ImpactIntervalEstimated   string
	ImpactStartOnEstimated    string
	ImpactEndOnEstimated      string
	AmountValidated           float64
	ImpactFrequencyValidated  string
	ImpactIntervalValidated   string
	ImpactStartOnValidated    string
	ImpactEndOnValidated      string
	GainShareFeePercentage    float64
	GainShareFeeStartOn       string
	GainShareFeeEndOn         string
	ProfitImprovementCategory string
	Organization              string
	OrganizationType          string
	OrganizationID            string
	Original                  string
	CreatedBy                 string
	OwnedBy                   string
	ValidatedBy               string
}

func newDoProfitImprovementsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a profit improvement",
		Long: `Create a profit improvement.

Required flags:
  --title                       Title
  --profit-improvement-category Category ID
  --organization                Organization in Type|ID format (e.g., Broker|123)

Optional flags:
  --description
  --status
  --amount-estimated
  --impact-frequency-estimated
  --impact-interval-estimated
  --impact-start-on-estimated
  --impact-end-on-estimated
  --amount-validated
  --impact-frequency-validated
  --impact-interval-validated
  --impact-start-on-validated
  --impact-end-on-validated
  --gain-share-fee-percentage
  --gain-share-fee-start-on
  --gain-share-fee-end-on
  --original
  --created-by
  --owned-by
  --validated-by
  --organization-type / --organization-id`,
		Example: `  # Create a profit improvement
  xbe do profit-improvements create \
    --title "Reduce idle time" \
    --profit-improvement-category 12 \
    --organization Broker|123

  # Create with estimated impact
  xbe do profit-improvements create \
    --title "Reduce idle time" \
    --profit-improvement-category 12 \
    --organization Broker|123 \
    --amount-estimated 5000 \
    --impact-frequency-estimated recurring \
    --impact-interval-estimated monthly`,
		RunE: runDoProfitImprovementsCreate,
	}
	initDoProfitImprovementsCreateFlags(cmd)
	return cmd
}

func init() {
	doProfitImprovementsCmd.AddCommand(newDoProfitImprovementsCreateCmd())
}

func initDoProfitImprovementsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("title", "", "Title (required)")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("status", "", "Status")
	cmd.Flags().Float64("amount-estimated", 0, "Estimated amount")
	cmd.Flags().String("impact-frequency-estimated", "", "Estimated impact frequency (one_time/recurring)")
	cmd.Flags().String("impact-interval-estimated", "", "Estimated impact interval (monthly/quarterly/annual)")
	cmd.Flags().String("impact-start-on-estimated", "", "Estimated impact start date (YYYY-MM-DD)")
	cmd.Flags().String("impact-end-on-estimated", "", "Estimated impact end date (YYYY-MM-DD)")
	cmd.Flags().Float64("amount-validated", 0, "Validated amount")
	cmd.Flags().String("impact-frequency-validated", "", "Validated impact frequency (one_time/recurring)")
	cmd.Flags().String("impact-interval-validated", "", "Validated impact interval (monthly/quarterly/annual)")
	cmd.Flags().String("impact-start-on-validated", "", "Validated impact start date (YYYY-MM-DD)")
	cmd.Flags().String("impact-end-on-validated", "", "Validated impact end date (YYYY-MM-DD)")
	cmd.Flags().Float64("gain-share-fee-percentage", 0, "Gain share fee percentage (0-1)")
	cmd.Flags().String("gain-share-fee-start-on", "", "Gain share fee start date (YYYY-MM-DD)")
	cmd.Flags().String("gain-share-fee-end-on", "", "Gain share fee end date (YYYY-MM-DD)")
	cmd.Flags().String("profit-improvement-category", "", "Profit improvement category ID (required)")
	cmd.Flags().String("organization", "", "Organization in Type|ID format (e.g., Broker|123)")
	cmd.Flags().String("organization-type", "", "Organization type (optional if --organization is set)")
	cmd.Flags().String("organization-id", "", "Organization ID (optional if --organization is set)")
	cmd.Flags().String("original", "", "Original profit improvement ID")
	cmd.Flags().String("created-by", "", "Created-by user ID")
	cmd.Flags().String("owned-by", "", "Owned-by user ID")
	cmd.Flags().String("validated-by", "", "Validated-by user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("title")
	_ = cmd.MarkFlagRequired("profit-improvement-category")
}

func runDoProfitImprovementsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProfitImprovementsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Title) == "" {
		err := fmt.Errorf("--title is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.ProfitImprovementCategory) == "" {
		err := fmt.Errorf("--profit-improvement-category is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	orgType, orgID, err := resolveProfitImprovementOrganization(cmd, opts.Organization, opts.OrganizationType, opts.OrganizationID)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if orgType == "" || orgID == "" {
		err := fmt.Errorf("--organization is required (format: Type|ID) or specify --organization-type and --organization-id")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"title": opts.Title,
	}
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("amount-estimated") {
		attributes["amount-estimated"] = opts.AmountEstimated
	}
	if cmd.Flags().Changed("impact-frequency-estimated") {
		attributes["impact-frequency-estimated"] = opts.ImpactFrequencyEstimated
	}
	if cmd.Flags().Changed("impact-interval-estimated") {
		attributes["impact-interval-estimated"] = opts.ImpactIntervalEstimated
	}
	if cmd.Flags().Changed("impact-start-on-estimated") {
		attributes["impact-start-on-estimated"] = opts.ImpactStartOnEstimated
	}
	if cmd.Flags().Changed("impact-end-on-estimated") {
		attributes["impact-end-on-estimated"] = opts.ImpactEndOnEstimated
	}
	if cmd.Flags().Changed("amount-validated") {
		attributes["amount-validated"] = opts.AmountValidated
	}
	if cmd.Flags().Changed("impact-frequency-validated") {
		attributes["impact-frequency-validated"] = opts.ImpactFrequencyValidated
	}
	if cmd.Flags().Changed("impact-interval-validated") {
		attributes["impact-interval-validated"] = opts.ImpactIntervalValidated
	}
	if cmd.Flags().Changed("impact-start-on-validated") {
		attributes["impact-start-on-validated"] = opts.ImpactStartOnValidated
	}
	if cmd.Flags().Changed("impact-end-on-validated") {
		attributes["impact-end-on-validated"] = opts.ImpactEndOnValidated
	}
	if cmd.Flags().Changed("gain-share-fee-percentage") {
		attributes["gain-share-fee-percentage"] = opts.GainShareFeePercentage
	}
	if cmd.Flags().Changed("gain-share-fee-start-on") {
		attributes["gain-share-fee-start-on"] = opts.GainShareFeeStartOn
	}
	if cmd.Flags().Changed("gain-share-fee-end-on") {
		attributes["gain-share-fee-end-on"] = opts.GainShareFeeEndOn
	}

	relationships := map[string]any{
		"profit-improvement-category": map[string]any{
			"data": map[string]any{
				"type": "profit-improvement-categories",
				"id":   opts.ProfitImprovementCategory,
			},
		},
		"organization": map[string]any{
			"data": map[string]any{
				"type": orgType,
				"id":   orgID,
			},
		},
	}

	if opts.Original != "" {
		relationships["original"] = map[string]any{
			"data": map[string]any{
				"type": "profit-improvements",
				"id":   opts.Original,
			},
		}
	}
	if opts.CreatedBy != "" {
		relationships["created-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.CreatedBy,
			},
		}
	}
	if opts.OwnedBy != "" {
		relationships["owned-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.OwnedBy,
			},
		}
	}
	if opts.ValidatedBy != "" {
		relationships["validated-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.ValidatedBy,
			},
		}
	}

	data := map[string]any{
		"type":          "profit-improvements",
		"attributes":    attributes,
		"relationships": relationships,
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

	body, _, err := client.Post(cmd.Context(), "/v1/profit-improvements", jsonBody)
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

	row := profitImprovementRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created profit improvement %s\n", row.ID)
	return nil
}

func parseDoProfitImprovementsCreateOptions(cmd *cobra.Command) (doProfitImprovementsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	title, _ := cmd.Flags().GetString("title")
	description, _ := cmd.Flags().GetString("description")
	status, _ := cmd.Flags().GetString("status")
	amountEstimated, _ := cmd.Flags().GetFloat64("amount-estimated")
	impactFrequencyEstimated, _ := cmd.Flags().GetString("impact-frequency-estimated")
	impactIntervalEstimated, _ := cmd.Flags().GetString("impact-interval-estimated")
	impactStartOnEstimated, _ := cmd.Flags().GetString("impact-start-on-estimated")
	impactEndOnEstimated, _ := cmd.Flags().GetString("impact-end-on-estimated")
	amountValidated, _ := cmd.Flags().GetFloat64("amount-validated")
	impactFrequencyValidated, _ := cmd.Flags().GetString("impact-frequency-validated")
	impactIntervalValidated, _ := cmd.Flags().GetString("impact-interval-validated")
	impactStartOnValidated, _ := cmd.Flags().GetString("impact-start-on-validated")
	impactEndOnValidated, _ := cmd.Flags().GetString("impact-end-on-validated")
	gainShareFeePercentage, _ := cmd.Flags().GetFloat64("gain-share-fee-percentage")
	gainShareFeeStartOn, _ := cmd.Flags().GetString("gain-share-fee-start-on")
	gainShareFeeEndOn, _ := cmd.Flags().GetString("gain-share-fee-end-on")
	profitImprovementCategory, _ := cmd.Flags().GetString("profit-improvement-category")
	organization, _ := cmd.Flags().GetString("organization")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	original, _ := cmd.Flags().GetString("original")
	createdBy, _ := cmd.Flags().GetString("created-by")
	ownedBy, _ := cmd.Flags().GetString("owned-by")
	validatedBy, _ := cmd.Flags().GetString("validated-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProfitImprovementsCreateOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		Title:                     title,
		Description:               description,
		Status:                    status,
		AmountEstimated:           amountEstimated,
		ImpactFrequencyEstimated:  impactFrequencyEstimated,
		ImpactIntervalEstimated:   impactIntervalEstimated,
		ImpactStartOnEstimated:    impactStartOnEstimated,
		ImpactEndOnEstimated:      impactEndOnEstimated,
		AmountValidated:           amountValidated,
		ImpactFrequencyValidated:  impactFrequencyValidated,
		ImpactIntervalValidated:   impactIntervalValidated,
		ImpactStartOnValidated:    impactStartOnValidated,
		ImpactEndOnValidated:      impactEndOnValidated,
		GainShareFeePercentage:    gainShareFeePercentage,
		GainShareFeeStartOn:       gainShareFeeStartOn,
		GainShareFeeEndOn:         gainShareFeeEndOn,
		ProfitImprovementCategory: profitImprovementCategory,
		Organization:              organization,
		OrganizationType:          organizationType,
		OrganizationID:            organizationID,
		Original:                  original,
		CreatedBy:                 createdBy,
		OwnedBy:                   ownedBy,
		ValidatedBy:               validatedBy,
	}, nil
}

func resolveProfitImprovementOrganization(cmd *cobra.Command, organization, orgType, orgID string) (string, string, error) {
	if cmd.Flags().Changed("organization") {
		return parseOrganization(organization)
	}
	if cmd.Flags().Changed("organization-type") || cmd.Flags().Changed("organization-id") {
		if strings.TrimSpace(orgType) == "" || strings.TrimSpace(orgID) == "" {
			return "", "", fmt.Errorf("--organization-type and --organization-id must be provided together")
		}
		return parseOrganization(fmt.Sprintf("%s|%s", orgType, orgID))
	}
	return "", "", nil
}
