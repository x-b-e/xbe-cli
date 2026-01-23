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

type doProjectCostCodesUpdateOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	ID                          string
	ExplicitCostCodeDescription string
}

func newDoProjectCostCodesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project cost code",
		Long: `Update a project cost code.

Optional:
  --explicit-description      Explicit cost code description`,
		Example: `  # Update explicit description
  xbe do project-cost-codes update 123 --explicit-description "Custom labor description"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectCostCodesUpdate,
	}
	initDoProjectCostCodesUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectCostCodesCmd.AddCommand(newDoProjectCostCodesUpdateCmd())
}

func initDoProjectCostCodesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("explicit-description", "", "Explicit cost code description")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectCostCodesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectCostCodesUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("explicit-description") {
		attributes["explicit-cost-code-description"] = opts.ExplicitCostCodeDescription
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "project-cost-codes",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/project-cost-codes/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project cost code %s\n", resp.Data.ID)
	return nil
}

func parseDoProjectCostCodesUpdateOptions(cmd *cobra.Command, args []string) (doProjectCostCodesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	explicitDescription, _ := cmd.Flags().GetString("explicit-description")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectCostCodesUpdateOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		ID:                          args[0],
		ExplicitCostCodeDescription: explicitDescription,
	}, nil
}
