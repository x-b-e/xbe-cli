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

type projectPhaseCostItemQuantityEstimatesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectPhaseCostItemQuantityEstimateDetails struct {
	ID                       string         `json:"id"`
	ProjectPhaseCostItemID   string         `json:"project_phase_cost_item_id,omitempty"`
	ProjectEstimateSetID     string         `json:"project_estimate_set_id,omitempty"`
	CreatedByID              string         `json:"created_by_id,omitempty"`
	RevenueItemQuantityBasis string         `json:"revenue_item_quantity_basis,omitempty"`
	Estimate                 map[string]any `json:"estimate,omitempty"`
	EstimateScaled           map[string]any `json:"estimate_scaled,omitempty"`
}

func newProjectPhaseCostItemQuantityEstimatesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project phase cost item quantity estimate details",
		Long: `Show the full details of a project phase cost item quantity estimate.

Output Fields:
  ID
  Project Phase Cost Item
  Project Estimate Set
  Created By
  Revenue Item Quantity Basis
  Estimate
  Estimate Scaled

Arguments:
  <id>  Quantity estimate ID (required). Use the list command to find IDs.`,
		Example: `  # Show a quantity estimate
  xbe view project-phase-cost-item-quantity-estimates show 123

  # Output as JSON
  xbe view project-phase-cost-item-quantity-estimates show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectPhaseCostItemQuantityEstimatesShow,
	}
	initProjectPhaseCostItemQuantityEstimatesShowFlags(cmd)
	return cmd
}

func init() {
	projectPhaseCostItemQuantityEstimatesCmd.AddCommand(newProjectPhaseCostItemQuantityEstimatesShowCmd())
}

func initProjectPhaseCostItemQuantityEstimatesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectPhaseCostItemQuantityEstimatesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseProjectPhaseCostItemQuantityEstimatesShowOptions(cmd)
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
		return fmt.Errorf("project phase cost item quantity estimate id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-phase-cost-item-quantity-estimates]", "revenue-item-quantity-basis,estimate,estimate-scaled,project-phase-cost-item,project-estimate-set,created-by")

	body, _, err := client.Get(cmd.Context(), "/v1/project-phase-cost-item-quantity-estimates/"+id, query)
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

	details := buildProjectPhaseCostItemQuantityEstimateDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectPhaseCostItemQuantityEstimateDetails(cmd, details)
}

func parseProjectPhaseCostItemQuantityEstimatesShowOptions(cmd *cobra.Command) (projectPhaseCostItemQuantityEstimatesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectPhaseCostItemQuantityEstimatesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectPhaseCostItemQuantityEstimateDetails(resp jsonAPISingleResponse) projectPhaseCostItemQuantityEstimateDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := projectPhaseCostItemQuantityEstimateDetails{
		ID:                       resource.ID,
		RevenueItemQuantityBasis: stringAttr(attrs, "revenue-item-quantity-basis"),
		Estimate:                 estimateAttr(attrs, "estimate"),
		EstimateScaled:           estimateAttr(attrs, "estimate-scaled"),
	}

	if rel, ok := resource.Relationships["project-phase-cost-item"]; ok && rel.Data != nil {
		details.ProjectPhaseCostItemID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["project-estimate-set"]; ok && rel.Data != nil {
		details.ProjectEstimateSetID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}

	return details
}

func renderProjectPhaseCostItemQuantityEstimateDetails(cmd *cobra.Command, details projectPhaseCostItemQuantityEstimateDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ProjectPhaseCostItemID != "" {
		fmt.Fprintf(out, "Project Phase Cost Item: %s\n", details.ProjectPhaseCostItemID)
	}
	if details.ProjectEstimateSetID != "" {
		fmt.Fprintf(out, "Project Estimate Set: %s\n", details.ProjectEstimateSetID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	if details.RevenueItemQuantityBasis != "" {
		fmt.Fprintf(out, "Revenue Item Quantity Basis: %s\n", details.RevenueItemQuantityBasis)
	}
	if formatted := formatAnyJSON(details.Estimate); formatted != "" {
		fmt.Fprintln(out, "Estimate:")
		fmt.Fprintln(out, formatted)
	}
	if formatted := formatAnyJSON(details.EstimateScaled); formatted != "" {
		fmt.Fprintln(out, "Estimate Scaled:")
		fmt.Fprintln(out, formatted)
	}

	return nil
}
