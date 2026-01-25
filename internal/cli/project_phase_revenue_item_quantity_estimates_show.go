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

type projectPhaseRevenueItemQuantityEstimatesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectPhaseRevenueItemQuantityEstimateDetails struct {
	ID                        string         `json:"id"`
	ProjectPhaseRevenueItemID string         `json:"project_phase_revenue_item_id,omitempty"`
	ProjectEstimateSetID      string         `json:"project_estimate_set_id,omitempty"`
	CreatedByID               string         `json:"created_by_id,omitempty"`
	Description               string         `json:"description,omitempty"`
	Estimate                  map[string]any `json:"estimate,omitempty"`
}

func newProjectPhaseRevenueItemQuantityEstimatesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project phase revenue item quantity estimate details",
		Long: `Show the full details of a project phase revenue item quantity estimate.

Output Fields:
  ID
  Project Phase Revenue Item
  Project Estimate Set
  Created By
  Description
  Estimate

Arguments:
  <id>  Quantity estimate ID (required). Use the list command to find IDs.`,
		Example: `  # Show a quantity estimate
  xbe view project-phase-revenue-item-quantity-estimates show 123

  # Output as JSON
  xbe view project-phase-revenue-item-quantity-estimates show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectPhaseRevenueItemQuantityEstimatesShow,
	}
	initProjectPhaseRevenueItemQuantityEstimatesShowFlags(cmd)
	return cmd
}

func init() {
	projectPhaseRevenueItemQuantityEstimatesCmd.AddCommand(newProjectPhaseRevenueItemQuantityEstimatesShowCmd())
}

func initProjectPhaseRevenueItemQuantityEstimatesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectPhaseRevenueItemQuantityEstimatesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectPhaseRevenueItemQuantityEstimatesShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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
		return fmt.Errorf("project phase revenue item quantity estimate id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-phase-revenue-item-quantity-estimates]", "description,estimate,project-phase-revenue-item,project-estimate-set,created-by")

	body, _, err := client.Get(cmd.Context(), "/v1/project-phase-revenue-item-quantity-estimates/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildProjectPhaseRevenueItemQuantityEstimateDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectPhaseRevenueItemQuantityEstimateDetails(cmd, details)
}

func parseProjectPhaseRevenueItemQuantityEstimatesShowOptions(cmd *cobra.Command) (projectPhaseRevenueItemQuantityEstimatesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectPhaseRevenueItemQuantityEstimatesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectPhaseRevenueItemQuantityEstimateDetails(resp jsonAPISingleResponse) projectPhaseRevenueItemQuantityEstimateDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := projectPhaseRevenueItemQuantityEstimateDetails{
		ID:          resource.ID,
		Description: stringAttr(attrs, "description"),
		Estimate:    estimateAttr(attrs, "estimate"),
	}

	if rel, ok := resource.Relationships["project-phase-revenue-item"]; ok && rel.Data != nil {
		details.ProjectPhaseRevenueItemID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["project-estimate-set"]; ok && rel.Data != nil {
		details.ProjectEstimateSetID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}

	return details
}

func renderProjectPhaseRevenueItemQuantityEstimateDetails(cmd *cobra.Command, details projectPhaseRevenueItemQuantityEstimateDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ProjectPhaseRevenueItemID != "" {
		fmt.Fprintf(out, "Project Phase Revenue Item: %s\n", details.ProjectPhaseRevenueItemID)
	}
	if details.ProjectEstimateSetID != "" {
		fmt.Fprintf(out, "Project Estimate Set: %s\n", details.ProjectEstimateSetID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	if details.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", details.Description)
	}
	if formatted := formatAnyJSON(details.Estimate); formatted != "" {
		fmt.Fprintln(out, "Estimate:")
		fmt.Fprintln(out, formatted)
	}

	return nil
}
