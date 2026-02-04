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

type projectPhaseCostItemPriceEstimatesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectPhaseCostItemPriceEstimateDetails struct {
	ID                     string `json:"id"`
	ProjectPhaseCostItemID string `json:"project_phase_cost_item_id,omitempty"`
	ProjectEstimateSetID   string `json:"project_estimate_set_id,omitempty"`
	CreatedByID            string `json:"created_by_id,omitempty"`
	CreatedAt              string `json:"created_at,omitempty"`
	UpdatedAt              string `json:"updated_at,omitempty"`
	Estimate               any    `json:"estimate,omitempty"`
}

func newProjectPhaseCostItemPriceEstimatesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project phase cost item price estimate details",
		Long: `Show the full details of a project phase cost item price estimate.

Output Fields:
  ID             Price estimate identifier
  Cost Item ID   Project phase cost item ID
  Estimate Set   Project estimate set ID
  Created By     Creator user ID
  Created At     Timestamp when the estimate was created
  Updated At     Timestamp when the estimate was last updated
  Estimate       Probability distribution (JSON)

Arguments:
  <id>    The price estimate ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a price estimate
  xbe view project-phase-cost-item-price-estimates show 123

  # JSON output
  xbe view project-phase-cost-item-price-estimates show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectPhaseCostItemPriceEstimatesShow,
	}
	initProjectPhaseCostItemPriceEstimatesShowFlags(cmd)
	return cmd
}

func init() {
	projectPhaseCostItemPriceEstimatesCmd.AddCommand(newProjectPhaseCostItemPriceEstimatesShowCmd())
}

func initProjectPhaseCostItemPriceEstimatesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectPhaseCostItemPriceEstimatesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseProjectPhaseCostItemPriceEstimatesShowOptions(cmd)
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
		return fmt.Errorf("price estimate id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-phase-cost-item-price-estimates]", "estimate,created-at,updated-at,project-phase-cost-item,project-estimate-set,created-by")

	body, _, err := client.Get(cmd.Context(), "/v1/project-phase-cost-item-price-estimates/"+id, query)
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

	details := buildProjectPhaseCostItemPriceEstimateDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectPhaseCostItemPriceEstimateDetails(cmd, details)
}

func parseProjectPhaseCostItemPriceEstimatesShowOptions(cmd *cobra.Command) (projectPhaseCostItemPriceEstimatesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectPhaseCostItemPriceEstimatesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectPhaseCostItemPriceEstimateDetails(resp jsonAPISingleResponse) projectPhaseCostItemPriceEstimateDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := projectPhaseCostItemPriceEstimateDetails{
		ID:                     resource.ID,
		Estimate:               attrs["estimate"],
		ProjectPhaseCostItemID: relationshipIDFromMap(resource.Relationships, "project-phase-cost-item"),
		ProjectEstimateSetID:   relationshipIDFromMap(resource.Relationships, "project-estimate-set"),
		CreatedByID:            relationshipIDFromMap(resource.Relationships, "created-by"),
		CreatedAt:              formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:              formatDateTime(stringAttr(attrs, "updated-at")),
	}

	return details
}

func renderProjectPhaseCostItemPriceEstimateDetails(cmd *cobra.Command, details projectPhaseCostItemPriceEstimateDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ProjectPhaseCostItemID != "" {
		fmt.Fprintf(out, "Cost Item ID: %s\n", details.ProjectPhaseCostItemID)
	}
	if details.ProjectEstimateSetID != "" {
		fmt.Fprintf(out, "Estimate Set ID: %s\n", details.ProjectEstimateSetID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By ID: %s\n", details.CreatedByID)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}
	if details.Estimate != nil {
		fmt.Fprintln(out, "\nEstimate:")
		fmt.Fprintln(out, formatJSONBlock(details.Estimate, "  "))
	}

	return nil
}
