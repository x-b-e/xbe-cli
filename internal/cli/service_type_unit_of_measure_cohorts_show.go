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

type serviceTypeUnitOfMeasureCohortsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type serviceTypeUnitOfMeasureCohortDetails struct {
	ID                          string   `json:"id"`
	Name                        string   `json:"name,omitempty"`
	IsActive                    bool     `json:"is_active"`
	CustomerID                  string   `json:"customer_id,omitempty"`
	Customer                    string   `json:"customer,omitempty"`
	TriggerID                   string   `json:"trigger_id,omitempty"`
	Trigger                     string   `json:"trigger,omitempty"`
	ServiceTypeUnitOfMeasureIDs []string `json:"service_type_unit_of_measure_ids,omitempty"`
}

func newServiceTypeUnitOfMeasureCohortsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show service type unit of measure cohort details",
		Long: `Show the full details of a service type unit of measure cohort.

Output Fields:
  ID
  Name
  Active
  Customer
  Trigger
  Service Type Unit of Measure IDs

Arguments:
  <id>    The cohort ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a cohort
  xbe view service-type-unit-of-measure-cohorts show 123

  # Output as JSON
  xbe view service-type-unit-of-measure-cohorts show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runServiceTypeUnitOfMeasureCohortsShow,
	}
	initServiceTypeUnitOfMeasureCohortsShowFlags(cmd)
	return cmd
}

func init() {
	serviceTypeUnitOfMeasureCohortsCmd.AddCommand(newServiceTypeUnitOfMeasureCohortsShowCmd())
}

func initServiceTypeUnitOfMeasureCohortsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runServiceTypeUnitOfMeasureCohortsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseServiceTypeUnitOfMeasureCohortsShowOptions(cmd)
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
		return fmt.Errorf("service type unit of measure cohort id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[service-type-unit-of-measure-cohorts]", "name,is-active,service-type-unit-of-measure-ids,customer,trigger")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[service-type-unit-of-measures]", "name")
	query.Set("include", "customer,trigger")

	body, _, err := client.Get(cmd.Context(), "/v1/service-type-unit-of-measure-cohorts/"+id, query)
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

	details := buildServiceTypeUnitOfMeasureCohortDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderServiceTypeUnitOfMeasureCohortDetails(cmd, details)
}

func parseServiceTypeUnitOfMeasureCohortsShowOptions(cmd *cobra.Command) (serviceTypeUnitOfMeasureCohortsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return serviceTypeUnitOfMeasureCohortsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildServiceTypeUnitOfMeasureCohortDetails(resp jsonAPISingleResponse) serviceTypeUnitOfMeasureCohortDetails {
	resource := resp.Data
	attrs := resource.Attributes
	included := map[string]map[string]any{}
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc.Attributes
	}

	details := serviceTypeUnitOfMeasureCohortDetails{
		ID:                          resource.ID,
		Name:                        stringAttr(attrs, "name"),
		IsActive:                    boolAttr(attrs, "is-active"),
		ServiceTypeUnitOfMeasureIDs: stringSliceAttr(attrs, "service-type-unit-of-measure-ids"),
	}

	details.CustomerID = relationshipIDFromMap(resource.Relationships, "customer")
	details.Customer = resolveServiceTypeUnitOfMeasureCohortCustomerName(details.CustomerID, included)
	details.TriggerID = relationshipIDFromMap(resource.Relationships, "trigger")
	details.Trigger = resolveServiceTypeUnitOfMeasureCohortTriggerName(details.TriggerID, included)

	return details
}

func renderServiceTypeUnitOfMeasureCohortDetails(cmd *cobra.Command, details serviceTypeUnitOfMeasureCohortDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}
	fmt.Fprintf(out, "Active: %t\n", details.IsActive)

	if details.CustomerID != "" {
		label := details.CustomerID
		if details.Customer != "" {
			label = fmt.Sprintf("%s (%s)", details.Customer, details.CustomerID)
		}
		fmt.Fprintf(out, "Customer: %s\n", label)
	}

	if details.TriggerID != "" {
		label := details.TriggerID
		if details.Trigger != "" {
			label = fmt.Sprintf("%s (%s)", details.Trigger, details.TriggerID)
		}
		fmt.Fprintf(out, "Trigger: %s\n", label)
	}

	if len(details.ServiceTypeUnitOfMeasureIDs) > 0 {
		fmt.Fprintf(out, "Service Type Unit of Measure IDs: %s\n", strings.Join(details.ServiceTypeUnitOfMeasureIDs, ", "))
	}

	return nil
}
