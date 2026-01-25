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

type doServiceTypeUnitOfMeasureCohortsUpdateOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	ID                          string
	Name                        string
	CustomerID                  string
	TriggerID                   string
	ServiceTypeUnitOfMeasureIDs []string
	IsActive                    bool
	NoActive                    bool
}

func newDoServiceTypeUnitOfMeasureCohortsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a service type unit of measure cohort",
		Long: `Update a service type unit of measure cohort.

Optional flags:
  --name                            Cohort name
  --service-type-unit-of-measure-ids Service type unit of measure IDs (comma-separated or repeated)
  --active                          Set as active
  --no-active                       Set as inactive
  --customer                        Customer ID
  --trigger                         Trigger service type unit of measure ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update cohort name
  xbe do service-type-unit-of-measure-cohorts update 123 --name "Updated cohort"

  # Update trigger and STUOM list
  xbe do service-type-unit-of-measure-cohorts update 123 \\
    --trigger 456 \\
    --service-type-unit-of-measure-ids 789,1011

  # Deactivate a cohort
  xbe do service-type-unit-of-measure-cohorts update 123 --no-active`,
		Args: cobra.ExactArgs(1),
		RunE: runDoServiceTypeUnitOfMeasureCohortsUpdate,
	}
	initDoServiceTypeUnitOfMeasureCohortsUpdateFlags(cmd)
	return cmd
}

func init() {
	doServiceTypeUnitOfMeasureCohortsCmd.AddCommand(newDoServiceTypeUnitOfMeasureCohortsUpdateCmd())
}

func initDoServiceTypeUnitOfMeasureCohortsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Cohort name")
	cmd.Flags().StringSlice("service-type-unit-of-measure-ids", nil, "Service type unit of measure IDs (comma-separated or repeated)")
	cmd.Flags().Bool("active", false, "Set as active")
	cmd.Flags().Bool("no-active", false, "Set as inactive")
	cmd.Flags().String("customer", "", "Customer ID")
	cmd.Flags().String("trigger", "", "Trigger service type unit of measure ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoServiceTypeUnitOfMeasureCohortsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoServiceTypeUnitOfMeasureCohortsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
	}
	if cmd.Flags().Changed("service-type-unit-of-measure-ids") {
		attributes["service-type-unit-of-measure-ids"] = opts.ServiceTypeUnitOfMeasureIDs
	}
	if cmd.Flags().Changed("active") {
		attributes["is-active"] = true
	}
	if cmd.Flags().Changed("no-active") {
		attributes["is-active"] = false
	}
	if cmd.Flags().Changed("customer") {
		relationships["customer"] = map[string]any{
			"data": map[string]any{
				"type": "customers",
				"id":   opts.CustomerID,
			},
		}
	}
	if cmd.Flags().Changed("trigger") {
		relationships["trigger"] = map[string]any{
			"data": map[string]any{
				"type": "service-type-unit-of-measures",
				"id":   opts.TriggerID,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "service-type-unit-of-measure-cohorts",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/service-type-unit-of-measure-cohorts/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated service type unit of measure cohort %s\n", row.ID)
	return nil
}

func parseDoServiceTypeUnitOfMeasureCohortsUpdateOptions(cmd *cobra.Command, args []string) (doServiceTypeUnitOfMeasureCohortsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	serviceTypeUnitOfMeasureIDs, _ := cmd.Flags().GetStringSlice("service-type-unit-of-measure-ids")
	isActive, _ := cmd.Flags().GetBool("active")
	noActive, _ := cmd.Flags().GetBool("no-active")
	customerID, _ := cmd.Flags().GetString("customer")
	triggerID, _ := cmd.Flags().GetString("trigger")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doServiceTypeUnitOfMeasureCohortsUpdateOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		ID:                          args[0],
		Name:                        name,
		CustomerID:                  customerID,
		TriggerID:                   triggerID,
		ServiceTypeUnitOfMeasureIDs: serviceTypeUnitOfMeasureIDs,
		IsActive:                    isActive,
		NoActive:                    noActive,
	}, nil
}
