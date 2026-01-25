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

type doServiceTypeUnitOfMeasureCohortsCreateOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	Name                        string
	IsActive                    bool
	CustomerID                  string
	TriggerID                   string
	ServiceTypeUnitOfMeasureIDs []string
}

func newDoServiceTypeUnitOfMeasureCohortsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new service type unit of measure cohort",
		Long: `Create a new service type unit of measure cohort.

Required flags:
  --customer                        Customer ID (required)
  --trigger                         Trigger service type unit of measure ID (required)
  --service-type-unit-of-measure-ids Service type unit of measure IDs (comma-separated or repeated) (required)

Optional flags:
  --name                            Cohort name
  --active                          Set as active (default: true)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a cohort
  xbe do service-type-unit-of-measure-cohorts create \\
    --customer 123 \\
    --trigger 456 \\
    --service-type-unit-of-measure-ids 789,1011

  # Create an inactive cohort
  xbe do service-type-unit-of-measure-cohorts create \\
    --customer 123 \\
    --trigger 456 \\
    --service-type-unit-of-measure-ids 789 \\
    --active=false`,
		Args: cobra.NoArgs,
		RunE: runDoServiceTypeUnitOfMeasureCohortsCreate,
	}
	initDoServiceTypeUnitOfMeasureCohortsCreateFlags(cmd)
	return cmd
}

func init() {
	doServiceTypeUnitOfMeasureCohortsCmd.AddCommand(newDoServiceTypeUnitOfMeasureCohortsCreateCmd())
}

func initDoServiceTypeUnitOfMeasureCohortsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Cohort name")
	cmd.Flags().Bool("active", true, "Set as active")
	cmd.Flags().String("customer", "", "Customer ID (required)")
	cmd.Flags().String("trigger", "", "Trigger service type unit of measure ID (required)")
	cmd.Flags().StringSlice("service-type-unit-of-measure-ids", nil, "Service type unit of measure IDs (comma-separated or repeated) (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoServiceTypeUnitOfMeasureCohortsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoServiceTypeUnitOfMeasureCohortsCreateOptions(cmd)
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

	if opts.CustomerID == "" {
		err := fmt.Errorf("--customer is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.TriggerID == "" {
		err := fmt.Errorf("--trigger is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if len(opts.ServiceTypeUnitOfMeasureIDs) == 0 {
		err := fmt.Errorf("--service-type-unit-of-measure-ids is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"service-type-unit-of-measure-ids": opts.ServiceTypeUnitOfMeasureIDs,
		"is-active":                        opts.IsActive,
	}

	if opts.Name != "" {
		attributes["name"] = opts.Name
	}

	relationships := map[string]any{
		"customer": map[string]any{
			"data": map[string]any{
				"type": "customers",
				"id":   opts.CustomerID,
			},
		},
		"trigger": map[string]any{
			"data": map[string]any{
				"type": "service-type-unit-of-measures",
				"id":   opts.TriggerID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "service-type-unit-of-measure-cohorts",
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

	body, _, err := client.Post(cmd.Context(), "/v1/service-type-unit-of-measure-cohorts", jsonBody)
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

	row := buildServiceTypeUnitOfMeasureCohortRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created service type unit of measure cohort %s\n", row.ID)
	return nil
}

func parseDoServiceTypeUnitOfMeasureCohortsCreateOptions(cmd *cobra.Command) (doServiceTypeUnitOfMeasureCohortsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	isActive, _ := cmd.Flags().GetBool("active")
	customerID, _ := cmd.Flags().GetString("customer")
	triggerID, _ := cmd.Flags().GetString("trigger")
	serviceTypeUnitOfMeasureIDs, _ := cmd.Flags().GetStringSlice("service-type-unit-of-measure-ids")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doServiceTypeUnitOfMeasureCohortsCreateOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		Name:                        name,
		IsActive:                    isActive,
		CustomerID:                  customerID,
		TriggerID:                   triggerID,
		ServiceTypeUnitOfMeasureIDs: serviceTypeUnitOfMeasureIDs,
	}, nil
}

func buildServiceTypeUnitOfMeasureCohortRowFromSingle(resp jsonAPISingleResponse) serviceTypeUnitOfMeasureCohortRow {
	included := map[string]map[string]any{}
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc.Attributes
	}
	return buildServiceTypeUnitOfMeasureCohortRow(resp.Data, included)
}
