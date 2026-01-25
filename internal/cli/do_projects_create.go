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

type doProjectsCreateOptions struct {
	BaseURL                                        string
	Token                                          string
	JSON                                           bool
	Name                                           string
	Number                                         string
	DueOn                                          string
	StartOn                                        string
	Developer                                      string
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
}

func newDoProjectsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new project",
		Long: `Create a new project.

Required flags:
  --name        The project name (required)
  --developer   The developer ID (required)

Optional flags:
  --number                                            Project number
  --due-on                                            Due date (ISO 8601, e.g. 2024-12-31)
  --start-on                                          Start date (ISO 8601, e.g. 2024-01-01)
  --project-manager                                   Project manager user ID
  --estimator                                         Estimator user ID
  --project-office                                    Project office ID
  --is-opportunity                                    Mark as opportunity (true/false)
  --is-inactive                                       Mark as inactive (true/false)
  --is-managed                                        Mark as managed (true/false)
  --is-prevailing-wage-explicit                       Prevailing wage (true/false)
  --is-certification-required-explicit                Certification required (true/false)
  --is-time-card-payroll-certification-required-explicit  Time card payroll certification required (true/false)
  --is-one-way-job-explicit                           One-way job (true/false)
  --is-transport-only                                 Transport only (true/false)
  --enforce-number-uniqueness                         Enforce number uniqueness (true/false)`,
		Example: `  # Create a project
  xbe do projects create --name "Highway 101" --developer 123

  # Create with dates
  xbe do projects create --name "Bridge Repair" --developer 123 --start-on 2024-01-01 --due-on 2024-06-30

  # Create with project manager
  xbe do projects create --name "Paving Job" --developer 123 --project-manager 456

  # Create as opportunity
  xbe do projects create --name "Potential Bid" --developer 123 --is-opportunity true

  # Create with prevailing wage requirement
  xbe do projects create --name "Government Job" --developer 123 --is-prevailing-wage-explicit true

  # Get JSON output
  xbe do projects create --name "New Project" --developer 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoProjectsCreate,
	}
	initDoProjectsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectsCmd.AddCommand(newDoProjectsCreateCmd())
}

func initDoProjectsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Project name (required)")
	cmd.Flags().String("developer", "", "Developer ID (required)")
	cmd.Flags().String("number", "", "Project number")
	cmd.Flags().String("due-on", "", "Due date (ISO 8601)")
	cmd.Flags().String("start-on", "", "Start date (ISO 8601)")
	cmd.Flags().String("project-manager", "", "Project manager user ID")
	cmd.Flags().String("estimator", "", "Estimator user ID")
	cmd.Flags().String("project-office", "", "Project office ID")
	cmd.Flags().String("is-opportunity", "", "Mark as opportunity (true/false)")
	cmd.Flags().String("is-inactive", "", "Mark as inactive (true/false)")
	cmd.Flags().String("is-managed", "", "Mark as managed (true/false)")
	cmd.Flags().String("is-prevailing-wage-explicit", "", "Prevailing wage (true/false)")
	cmd.Flags().String("is-certification-required-explicit", "", "Certification required (true/false)")
	cmd.Flags().String("is-time-card-payroll-certification-required-explicit", "", "Time card payroll certification required (true/false)")
	cmd.Flags().String("is-one-way-job-explicit", "", "One-way job (true/false)")
	cmd.Flags().String("is-transport-only", "", "Transport only (true/false)")
	cmd.Flags().String("enforce-number-uniqueness", "", "Enforce number uniqueness (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectsCreateOptions(cmd)
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

	// Require name
	if opts.Name == "" {
		err := fmt.Errorf("--name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require developer
	if opts.Developer == "" {
		err := fmt.Errorf("--developer is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Build attributes
	attributes := map[string]any{
		"name": opts.Name,
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

	// Build relationships
	relationships := map[string]any{
		"developer": map[string]any{
			"data": map[string]string{
				"type": "developers",
				"id":   opts.Developer,
			},
		},
	}
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

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "projects",
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

	body, _, err := client.Post(cmd.Context(), "/v1/projects", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created project %s (%s)\n", result["id"], result["name"])
	return nil
}

func parseDoProjectsCreateOptions(cmd *cobra.Command) (doProjectsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	developer, _ := cmd.Flags().GetString("developer")
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
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectsCreateOptions{
		BaseURL:                         baseURL,
		Token:                           token,
		JSON:                            jsonOut,
		Name:                            name,
		Developer:                       developer,
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
	}, nil
}
