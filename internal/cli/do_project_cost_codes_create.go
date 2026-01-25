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

type doProjectCostCodesCreateOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	ProjectCustomer             string
	CostCode                    string
	ExplicitCostCodeDescription string
}

func newDoProjectCostCodesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project cost code",
		Long: `Create a project cost code.

Required:
  --project-customer          Project customer ID
  --cost-code                 Cost code ID

Optional:
  --explicit-description      Explicit cost code description`,
		Example: `  # Create a project cost code
  xbe do project-cost-codes create --project-customer 123 --cost-code 456

  # Create with explicit description
  xbe do project-cost-codes create --project-customer 123 --cost-code 456 --explicit-description "Custom labor"`,
		RunE: runDoProjectCostCodesCreate,
	}
	initDoProjectCostCodesCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectCostCodesCmd.AddCommand(newDoProjectCostCodesCreateCmd())
}

func initDoProjectCostCodesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-customer", "", "Project customer ID")
	cmd.Flags().String("cost-code", "", "Cost code ID")
	cmd.Flags().String("explicit-description", "", "Explicit cost code description")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("project-customer")
	_ = cmd.MarkFlagRequired("cost-code")
}

func runDoProjectCostCodesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectCostCodesCreateOptions(cmd)
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

	if opts.ExplicitCostCodeDescription != "" {
		attributes["explicit-cost-code-description"] = opts.ExplicitCostCodeDescription
	}

	relationships := map[string]any{
		"project-customer": map[string]any{
			"data": map[string]any{
				"type": "project-customers",
				"id":   opts.ProjectCustomer,
			},
		},
		"cost-code": map[string]any{
			"data": map[string]any{
				"type": "cost-codes",
				"id":   opts.CostCode,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-cost-codes",
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-cost-codes", jsonBody)
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

	if opts.JSON {
		row := projectCostCodeRow{
			ID:                          resp.Data.ID,
			ExplicitCostCodeDescription: stringAttr(resp.Data.Attributes, "explicit-cost-code-description"),
			CostCodeDescription:         stringAttr(resp.Data.Attributes, "cost-code-description"),
		}
		if rel, ok := resp.Data.Relationships["project-customer"]; ok && rel.Data != nil {
			row.ProjectCustomerID = rel.Data.ID
		}
		if rel, ok := resp.Data.Relationships["cost-code"]; ok && rel.Data != nil {
			row.CostCodeID = rel.Data.ID
		}
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project cost code %s\n", resp.Data.ID)
	return nil
}

func parseDoProjectCostCodesCreateOptions(cmd *cobra.Command) (doProjectCostCodesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectCustomer, _ := cmd.Flags().GetString("project-customer")
	costCode, _ := cmd.Flags().GetString("cost-code")
	explicitDescription, _ := cmd.Flags().GetString("explicit-description")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectCostCodesCreateOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		ProjectCustomer:             projectCustomer,
		CostCode:                    costCode,
		ExplicitCostCodeDescription: explicitDescription,
	}, nil
}
