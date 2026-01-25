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

type doProfitImprovementsUpdateOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	ID                        string
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

func newDoProfitImprovementsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a profit improvement",
		Long: `Update an existing profit improvement.

Optional flags:
  --title
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
  --profit-improvement-category
  --organization / --organization-type / --organization-id
  --original
  --created-by
  --owned-by
  --validated-by`,
		Example: `  # Update a profit improvement title
  xbe do profit-improvements update 123 --title "Updated title"

  # Update estimated impact
  xbe do profit-improvements update 123 \
    --amount-estimated 2500 \
    --impact-frequency-estimated recurring \
    --impact-interval-estimated monthly`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProfitImprovementsUpdate,
	}
	initDoProfitImprovementsUpdateFlags(cmd)
	return cmd
}

func init() {
	doProfitImprovementsCmd.AddCommand(newDoProfitImprovementsUpdateCmd())
}

func initDoProfitImprovementsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("title", "", "Title")
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
	cmd.Flags().String("profit-improvement-category", "", "Profit improvement category ID")
	cmd.Flags().String("organization", "", "Organization in Type|ID format (e.g., Broker|123)")
	cmd.Flags().String("organization-type", "", "Organization type")
	cmd.Flags().String("organization-id", "", "Organization ID")
	cmd.Flags().String("original", "", "Original profit improvement ID")
	cmd.Flags().String("created-by", "", "Created-by user ID")
	cmd.Flags().String("owned-by", "", "Owned-by user ID")
	cmd.Flags().String("validated-by", "", "Validated-by user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProfitImprovementsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProfitImprovementsUpdateOptions(cmd, args)
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
	relationships := map[string]any{}

	if cmd.Flags().Changed("title") {
		attributes["title"] = opts.Title
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

	if cmd.Flags().Changed("profit-improvement-category") {
		relationships["profit-improvement-category"] = map[string]any{
			"data": map[string]any{
				"type": "profit-improvement-categories",
				"id":   opts.ProfitImprovementCategory,
			},
		}
	}
	if cmd.Flags().Changed("original") {
		relationships["original"] = map[string]any{
			"data": map[string]any{
				"type": "profit-improvements",
				"id":   opts.Original,
			},
		}
	}
	if cmd.Flags().Changed("created-by") {
		relationships["created-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.CreatedBy,
			},
		}
	}
	if cmd.Flags().Changed("owned-by") {
		relationships["owned-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.OwnedBy,
			},
		}
	}
	if cmd.Flags().Changed("validated-by") {
		relationships["validated-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.ValidatedBy,
			},
		}
	}

	orgType, orgID, err := resolveProfitImprovementOrganization(cmd, opts.Organization, opts.OrganizationType, opts.OrganizationID)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if orgType != "" && orgID != "" {
		relationships["organization"] = map[string]any{
			"data": map[string]any{
				"type": orgType,
				"id":   orgID,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := errors.New("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "profit-improvements",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
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

	path := fmt.Sprintf("/v1/profit-improvements/%s", opts.ID)
	body, _, err := client.Patch(cmd.Context(), path, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated profit improvement %s\n", row.ID)
	return nil
}

func parseDoProfitImprovementsUpdateOptions(cmd *cobra.Command, args []string) (doProfitImprovementsUpdateOptions, error) {
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

	return doProfitImprovementsUpdateOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		ID:                        args[0],
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
