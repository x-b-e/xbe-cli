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

type serviceTypeUnitOfMeasuresShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type serviceTypeUnitOfMeasureDetails struct {
	ID                                string `json:"id"`
	Name                              string `json:"name,omitempty"`
	ServiceTypeID                     string `json:"service_type_id,omitempty"`
	ServiceType                       string `json:"service_type,omitempty"`
	UnitOfMeasureID                   string `json:"unit_of_measure_id,omitempty"`
	UnitOfMeasure                     string `json:"unit_of_measure,omitempty"`
	CommonDenominatorID               string `json:"common_denominator_id,omitempty"`
	CommonDenominator                 string `json:"common_denominator,omitempty"`
	IsQuantifiable                    bool   `json:"is_quantifiable"`
	IsSupplementalQuantifiable        bool   `json:"is_supplemental_quantifiable"`
	UnitOfMeasureCalculationType      string `json:"unit_of_measure_calculation_type,omitempty"`
	JobCoefficientMethod              string `json:"job_coefficient_method,omitempty"`
	AreQuantitiesCreatedAutomatically bool   `json:"are_quantities_created_automatically"`
	UserCanAssign                     bool   `json:"user_can_assign"`
	IsQuantitySetBeforeValidation     bool   `json:"is_quantity_set_before_validation"`
	ReasonableDefaultQuantity         string `json:"reasonable_default_quantity,omitempty"`
	MinimumReasonableQuantity         string `json:"minimum_reasonable_quantity,omitempty"`
	MaximumReasonableQuantity         string `json:"maximum_reasonable_quantity,omitempty"`
}

func newServiceTypeUnitOfMeasuresShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show service type unit of measure details",
		Long: `Show the full details of a service type unit of measure.

Output Fields:
  ID
  Name
  Service Type (ID + name)
  Unit of Measure (ID + name)
  Common Denominator (ID + name)
  Quantifiable
  Supplemental Quantifiable
  Unit of Measure Calculation Type
  Job Coefficient Method
  Are Quantities Created Automatically
  User Can Assign
  Is Quantity Set Before Validation
  Reasonable Default Quantity
  Minimum Reasonable Quantity
  Maximum Reasonable Quantity

Arguments:
  <id>    The service type unit of measure ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a service type unit of measure
  xbe view service-type-unit-of-measures show 123

  # Output as JSON
  xbe view service-type-unit-of-measures show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runServiceTypeUnitOfMeasuresShow,
	}
	initServiceTypeUnitOfMeasuresShowFlags(cmd)
	return cmd
}

func init() {
	serviceTypeUnitOfMeasuresCmd.AddCommand(newServiceTypeUnitOfMeasuresShowCmd())
}

func initServiceTypeUnitOfMeasuresShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runServiceTypeUnitOfMeasuresShow(cmd *cobra.Command, args []string) error {
	opts, err := parseServiceTypeUnitOfMeasuresShowOptions(cmd)
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
		return fmt.Errorf("service type unit of measure id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[service-type-unit-of-measures]", "name,is-quantifiable,is-supplemental-quantifiable,unit-of-measure-calculation-type,job-coefficient-method,are-quantities-created-automatically,user-can-assign,is-quantity-set-before-validation,reasonable-default-quantity,minimum-reasonable-quantity,maximum-reasonable-quantity,service-type,unit-of-measure,common-denominator")
	query.Set("fields[service-types]", "name")
	query.Set("fields[unit-of-measures]", "name,abbreviation")
	query.Set("include", "service-type,unit-of-measure,common-denominator")

	body, _, err := client.Get(cmd.Context(), "/v1/service-type-unit-of-measures/"+id, query)
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

	details := buildServiceTypeUnitOfMeasureDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderServiceTypeUnitOfMeasureDetails(cmd, details)
}

func parseServiceTypeUnitOfMeasuresShowOptions(cmd *cobra.Command) (serviceTypeUnitOfMeasuresShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return serviceTypeUnitOfMeasuresShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildServiceTypeUnitOfMeasureDetails(resp jsonAPISingleResponse) serviceTypeUnitOfMeasureDetails {
	resource := resp.Data
	attrs := resource.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := serviceTypeUnitOfMeasureDetails{
		ID:                                resource.ID,
		Name:                              stringAttr(attrs, "name"),
		IsQuantifiable:                    boolAttr(attrs, "is-quantifiable"),
		IsSupplementalQuantifiable:        boolAttr(attrs, "is-supplemental-quantifiable"),
		UnitOfMeasureCalculationType:      stringAttr(attrs, "unit-of-measure-calculation-type"),
		JobCoefficientMethod:              stringAttr(attrs, "job-coefficient-method"),
		AreQuantitiesCreatedAutomatically: boolAttr(attrs, "are-quantities-created-automatically"),
		UserCanAssign:                     boolAttr(attrs, "user-can-assign"),
		IsQuantitySetBeforeValidation:     boolAttr(attrs, "is-quantity-set-before-validation"),
		ReasonableDefaultQuantity:         stringAttr(attrs, "reasonable-default-quantity"),
		MinimumReasonableQuantity:         stringAttr(attrs, "minimum-reasonable-quantity"),
		MaximumReasonableQuantity:         stringAttr(attrs, "maximum-reasonable-quantity"),
	}

	if rel, ok := resource.Relationships["service-type"]; ok && rel.Data != nil {
		details.ServiceTypeID = rel.Data.ID
		if serviceType, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ServiceType = stringAttr(serviceType.Attributes, "name")
		}
	}

	if rel, ok := resource.Relationships["unit-of-measure"]; ok && rel.Data != nil {
		details.UnitOfMeasureID = rel.Data.ID
		if uom, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.UnitOfMeasure = firstNonEmpty(
				stringAttr(uom.Attributes, "abbreviation"),
				stringAttr(uom.Attributes, "name"),
			)
		}
	}

	if rel, ok := resource.Relationships["common-denominator"]; ok && rel.Data != nil {
		details.CommonDenominatorID = rel.Data.ID
		if stuom, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.CommonDenominator = stringAttr(stuom.Attributes, "name")
		}
	}

	return details
}

func renderServiceTypeUnitOfMeasureDetails(cmd *cobra.Command, details serviceTypeUnitOfMeasureDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}
	if details.ServiceType != "" {
		fmt.Fprintf(out, "Service Type: %s\n", details.ServiceType)
	}
	if details.ServiceTypeID != "" {
		fmt.Fprintf(out, "Service Type ID: %s\n", details.ServiceTypeID)
	}
	if details.UnitOfMeasure != "" {
		fmt.Fprintf(out, "Unit of Measure: %s\n", details.UnitOfMeasure)
	}
	if details.UnitOfMeasureID != "" {
		fmt.Fprintf(out, "Unit of Measure ID: %s\n", details.UnitOfMeasureID)
	}
	if details.CommonDenominator != "" {
		fmt.Fprintf(out, "Common Denominator: %s\n", details.CommonDenominator)
	}
	if details.CommonDenominatorID != "" {
		fmt.Fprintf(out, "Common Denominator ID: %s\n", details.CommonDenominatorID)
	}
	fmt.Fprintf(out, "Quantifiable: %t\n", details.IsQuantifiable)
	fmt.Fprintf(out, "Supplemental Quantifiable: %t\n", details.IsSupplementalQuantifiable)
	if details.UnitOfMeasureCalculationType != "" {
		fmt.Fprintf(out, "Unit of Measure Calculation Type: %s\n", details.UnitOfMeasureCalculationType)
	}
	if details.JobCoefficientMethod != "" {
		fmt.Fprintf(out, "Job Coefficient Method: %s\n", details.JobCoefficientMethod)
	}
	fmt.Fprintf(out, "Are Quantities Created Automatically: %t\n", details.AreQuantitiesCreatedAutomatically)
	fmt.Fprintf(out, "User Can Assign: %t\n", details.UserCanAssign)
	fmt.Fprintf(out, "Is Quantity Set Before Validation: %t\n", details.IsQuantitySetBeforeValidation)
	if details.ReasonableDefaultQuantity != "" {
		fmt.Fprintf(out, "Reasonable Default Quantity: %s\n", details.ReasonableDefaultQuantity)
	}
	if details.MinimumReasonableQuantity != "" {
		fmt.Fprintf(out, "Minimum Reasonable Quantity: %s\n", details.MinimumReasonableQuantity)
	}
	if details.MaximumReasonableQuantity != "" {
		fmt.Fprintf(out, "Maximum Reasonable Quantity: %s\n", details.MaximumReasonableQuantity)
	}

	return nil
}
