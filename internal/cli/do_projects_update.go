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

type doProjectsUpdateOptions struct {
	BaseURL                                        string
	Token                                          string
	JSON                                           bool
	Name                                           string
	Number                                         string
	DueOn                                          string
	StartOn                                        string
	ProjectManager                                 string
	Estimator                                      string
	ProjectOffice                                  string
	IsOpportunity                                  string
	IsInactive                                     string
	IsManaged                                      string
	IsPrevailingWageExplicit                       string
	IsCertificationRequiredExplicit                string
	IsTimeCardPayrollCertificationRequiredExplicit string
	IsOneWayJobExplicit                            string
	IsTransportOnly                                string
	EnforceNumberUniqueness                        string
	BidEstimateSet                                 string
	ActualEstimateSet                              string
	PossibleEstimateSet                            string
	ProjectTransportPlan                           string
}

func newDoProjectsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project",
		Long: `Update an existing project.

Only the fields you specify will be updated. Fields not provided will remain unchanged.

Arguments:
  <id>    The project ID (required)

Flags:
  --name                                              Update the name
  --number                                            Update the project number
  --due-on                                            Update due date (ISO 8601)
  --start-on                                          Update start date (ISO 8601)
  --project-manager                                   Update project manager user ID
  --estimator                                         Update estimator user ID
  --project-office                                    Update project office ID
  --is-opportunity                                    Update opportunity status (true/false)
  --is-inactive                                       Update inactive status (true/false)
  --is-managed                                        Update managed status (true/false)
  --is-prevailing-wage-explicit                       Prevailing wage (true/false)
  --is-certification-required-explicit                Certification required (true/false)
  --is-time-card-payroll-certification-required-explicit  Time card payroll certification required (true/false)
  --is-one-way-job-explicit                           One-way job (true/false)
  --is-transport-only                                 Transport only (true/false)
  --enforce-number-uniqueness                         Enforce number uniqueness (true/false)
  --bid-estimate-set                                  Bid estimate set ID
  --actual-estimate-set                               Actual estimate set ID
  --possible-estimate-set                             Possible estimate set ID
  --project-transport-plan                            Project transport plan ID`,
		Example: `  # Update the name
  xbe do projects update 123 --name "New Project Name"

  # Update dates
  xbe do projects update 123 --due-on 2024-12-31

  # Update project manager
  xbe do projects update 123 --project-manager 456

  # Mark as inactive
  xbe do projects update 123 --is-inactive true

  # Set prevailing wage requirement
  xbe do projects update 123 --is-prevailing-wage-explicit true

  # Set bid estimate set
  xbe do projects update 123 --bid-estimate-set 789

  # Get JSON output
  xbe do projects update 123 --name "Updated" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectsUpdate,
	}
	initDoProjectsUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectsCmd.AddCommand(newDoProjectsUpdateCmd())
}

func initDoProjectsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "New name")
	cmd.Flags().String("number", "", "New project number")
	cmd.Flags().String("due-on", "", "New due date (ISO 8601)")
	cmd.Flags().String("start-on", "", "New start date (ISO 8601)")
	cmd.Flags().String("project-manager", "", "New project manager user ID")
	cmd.Flags().String("estimator", "", "New estimator user ID")
	cmd.Flags().String("project-office", "", "New project office ID")
	cmd.Flags().String("is-opportunity", "", "Update opportunity status (true/false)")
	cmd.Flags().String("is-inactive", "", "Update inactive status (true/false)")
	cmd.Flags().String("is-managed", "", "Update managed status (true/false)")
	cmd.Flags().String("is-prevailing-wage-explicit", "", "Prevailing wage (true/false)")
	cmd.Flags().String("is-certification-required-explicit", "", "Certification required (true/false)")
	cmd.Flags().String("is-time-card-payroll-certification-required-explicit", "", "Time card payroll certification required (true/false)")
	cmd.Flags().String("is-one-way-job-explicit", "", "One-way job (true/false)")
	cmd.Flags().String("is-transport-only", "", "Transport only (true/false)")
	cmd.Flags().String("enforce-number-uniqueness", "", "Enforce number uniqueness (true/false)")
	cmd.Flags().String("bid-estimate-set", "", "Bid estimate set ID")
	cmd.Flags().String("actual-estimate-set", "", "Actual estimate set ID")
	cmd.Flags().String("possible-estimate-set", "", "Possible estimate set ID")
	cmd.Flags().String("project-transport-plan", "", "Project transport plan ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectsUpdateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require authentication for write operations
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("project id is required")
	}

	// Require at least one field to update
	if opts.Name == "" && opts.Number == "" && opts.DueOn == "" && opts.StartOn == "" &&
		opts.ProjectManager == "" && opts.Estimator == "" && opts.ProjectOffice == "" &&
		opts.IsOpportunity == "" && opts.IsInactive == "" && opts.IsManaged == "" &&
		opts.IsPrevailingWageExplicit == "" && opts.IsCertificationRequiredExplicit == "" &&
		opts.IsTimeCardPayrollCertificationRequiredExplicit == "" && opts.IsOneWayJobExplicit == "" &&
		opts.IsTransportOnly == "" && opts.EnforceNumberUniqueness == "" &&
		opts.BidEstimateSet == "" && opts.ActualEstimateSet == "" &&
		opts.PossibleEstimateSet == "" && opts.ProjectTransportPlan == "" {
		err := fmt.Errorf("at least one field to update is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Build attributes
	attributes := map[string]any{}
	if opts.Name != "" {
		attributes["name"] = opts.Name
	}
	if opts.Number != "" {
		attributes["number"] = opts.Number
	}
	if opts.DueOn != "" {
		attributes["due-on"] = opts.DueOn
	}
	if opts.StartOn != "" {
		attributes["start-on"] = opts.StartOn
	}
	if opts.IsOpportunity != "" {
		attributes["is-opportunity"] = opts.IsOpportunity == "true"
	}
	if opts.IsInactive != "" {
		attributes["is-inactive-explicit"] = opts.IsInactive == "true"
	}
	if opts.IsManaged != "" {
		attributes["is-managed"] = opts.IsManaged == "true"
	}
	if opts.IsPrevailingWageExplicit != "" {
		attributes["is-prevailing-wage-explicit"] = opts.IsPrevailingWageExplicit == "true"
	}
	if opts.IsCertificationRequiredExplicit != "" {
		attributes["is-certification-required-explicit"] = opts.IsCertificationRequiredExplicit == "true"
	}
	if opts.IsTimeCardPayrollCertificationRequiredExplicit != "" {
		attributes["is-time-card-payroll-certification-required-explicit"] = opts.IsTimeCardPayrollCertificationRequiredExplicit == "true"
	}
	if opts.IsOneWayJobExplicit != "" {
		attributes["is-one-way-job-explicit"] = opts.IsOneWayJobExplicit == "true"
	}
	if opts.IsTransportOnly != "" {
		attributes["is-transport-only"] = opts.IsTransportOnly == "true"
	}
	if opts.EnforceNumberUniqueness != "" {
		attributes["enforce-number-uniqueness"] = opts.EnforceNumberUniqueness == "true"
	}

	data := map[string]any{
		"id":         id,
		"type":       "projects",
		"attributes": attributes,
	}

	// Build relationships
	relationships := map[string]any{}
	if opts.ProjectManager != "" {
		relationships["project-manager"] = map[string]any{
			"data": map[string]string{
				"type": "users",
				"id":   opts.ProjectManager,
			},
		}
	}
	if opts.Estimator != "" {
		relationships["estimator"] = map[string]any{
			"data": map[string]string{
				"type": "users",
				"id":   opts.Estimator,
			},
		}
	}
	if opts.ProjectOffice != "" {
		relationships["project-office"] = map[string]any{
			"data": map[string]string{
				"type": "project-offices",
				"id":   opts.ProjectOffice,
			},
		}
	}
	if opts.BidEstimateSet != "" {
		relationships["bid-estimate-set"] = map[string]any{
			"data": map[string]string{
				"type": "project-estimate-sets",
				"id":   opts.BidEstimateSet,
			},
		}
	}
	if opts.ActualEstimateSet != "" {
		relationships["actual-estimate-set"] = map[string]any{
			"data": map[string]string{
				"type": "project-estimate-sets",
				"id":   opts.ActualEstimateSet,
			},
		}
	}
	if opts.PossibleEstimateSet != "" {
		relationships["possible-estimate-set"] = map[string]any{
			"data": map[string]string{
				"type": "project-estimate-sets",
				"id":   opts.PossibleEstimateSet,
			},
		}
	}
	if opts.ProjectTransportPlan != "" {
		relationships["project-transport-plan"] = map[string]any{
			"data": map[string]string{
				"type": "project-transport-plans",
				"id":   opts.ProjectTransportPlan,
			},
		}
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

	body, _, err := client.Patch(cmd.Context(), "/v1/projects/"+id, jsonBody)
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

	result := map[string]any{
		"id":     resp.Data.ID,
		"name":   stringAttr(resp.Data.Attributes, "name"),
		"number": stringAttr(resp.Data.Attributes, "number"),
		"status": stringAttr(resp.Data.Attributes, "status"),
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), result)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project %s (%s)\n", result["id"], result["name"])
	return nil
}

func parseDoProjectsUpdateOptions(cmd *cobra.Command) (doProjectsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	number, _ := cmd.Flags().GetString("number")
	dueOn, _ := cmd.Flags().GetString("due-on")
	startOn, _ := cmd.Flags().GetString("start-on")
	projectManager, _ := cmd.Flags().GetString("project-manager")
	estimator, _ := cmd.Flags().GetString("estimator")
	projectOffice, _ := cmd.Flags().GetString("project-office")
	isOpportunity, _ := cmd.Flags().GetString("is-opportunity")
	isInactive, _ := cmd.Flags().GetString("is-inactive")
	isManaged, _ := cmd.Flags().GetString("is-managed")
	isPrevailingWageExplicit, _ := cmd.Flags().GetString("is-prevailing-wage-explicit")
	isCertificationRequiredExplicit, _ := cmd.Flags().GetString("is-certification-required-explicit")
	isTimeCardPayrollCertificationRequiredExplicit, _ := cmd.Flags().GetString("is-time-card-payroll-certification-required-explicit")
	isOneWayJobExplicit, _ := cmd.Flags().GetString("is-one-way-job-explicit")
	isTransportOnly, _ := cmd.Flags().GetString("is-transport-only")
	enforceNumberUniqueness, _ := cmd.Flags().GetString("enforce-number-uniqueness")
	bidEstimateSet, _ := cmd.Flags().GetString("bid-estimate-set")
	actualEstimateSet, _ := cmd.Flags().GetString("actual-estimate-set")
	possibleEstimateSet, _ := cmd.Flags().GetString("possible-estimate-set")
	projectTransportPlan, _ := cmd.Flags().GetString("project-transport-plan")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectsUpdateOptions{
		BaseURL:                         baseURL,
		Token:                           token,
		JSON:                            jsonOut,
		Name:                            name,
		Number:                          number,
		DueOn:                           dueOn,
		StartOn:                         startOn,
		ProjectManager:                  projectManager,
		Estimator:                       estimator,
		ProjectOffice:                   projectOffice,
		IsOpportunity:                   isOpportunity,
		IsInactive:                      isInactive,
		IsManaged:                       isManaged,
		IsPrevailingWageExplicit:        isPrevailingWageExplicit,
		IsCertificationRequiredExplicit: isCertificationRequiredExplicit,
		IsTimeCardPayrollCertificationRequiredExplicit: isTimeCardPayrollCertificationRequiredExplicit,
		IsOneWayJobExplicit:                            isOneWayJobExplicit,
		IsTransportOnly:                                isTransportOnly,
		EnforceNumberUniqueness:                        enforceNumberUniqueness,
		BidEstimateSet:                                 bidEstimateSet,
		ActualEstimateSet:                              actualEstimateSet,
		PossibleEstimateSet:                            possibleEstimateSet,
		ProjectTransportPlan:                           projectTransportPlan,
	}, nil
}
