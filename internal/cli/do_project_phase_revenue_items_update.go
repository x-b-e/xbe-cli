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

type doProjectPhaseRevenueItemsUpdateOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	ID                           string
	ProjectRevenueClassification string
	QuantityStrategy             string
	Note                         string
	QuantityEstimate             string
}

func newDoProjectPhaseRevenueItemsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project phase revenue item",
		Long: `Update an existing project phase revenue item.

Only the fields you specify will be updated. Fields not provided will remain unchanged.

Arguments:
  <id>    The project phase revenue item ID (required)

Flags:
  --project-revenue-classification Update project revenue classification ID (empty to clear)
  --quantity-strategy              Update quantity strategy (direct/indirect)
  --note                           Update note (empty to clear)
  --quantity-estimate              Update quantity estimate ID (empty to clear)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update quantity strategy
  xbe do project-phase-revenue-items update 123 --quantity-strategy indirect

  # Update note
  xbe do project-phase-revenue-items update 123 --note "Revised note"

  # Clear project revenue classification
  xbe do project-phase-revenue-items update 123 --project-revenue-classification ""`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectPhaseRevenueItemsUpdate,
	}
	initDoProjectPhaseRevenueItemsUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectPhaseRevenueItemsCmd.AddCommand(newDoProjectPhaseRevenueItemsUpdateCmd())
}

func initDoProjectPhaseRevenueItemsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-revenue-classification", "", "Project revenue classification ID")
	cmd.Flags().String("quantity-strategy", "", "Quantity strategy (direct/indirect)")
	cmd.Flags().String("note", "", "Note")
	cmd.Flags().String("quantity-estimate", "", "Quantity estimate ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectPhaseRevenueItemsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectPhaseRevenueItemsUpdateOptions(cmd, args)
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
	hasChanges := false

	if cmd.Flags().Changed("quantity-strategy") {
		attributes["quantity-strategy"] = opts.QuantityStrategy
		hasChanges = true
	}

	if cmd.Flags().Changed("note") {
		attributes["note"] = opts.Note
		hasChanges = true
	}

	if cmd.Flags().Changed("project-revenue-classification") {
		if opts.ProjectRevenueClassification == "" {
			relationships["project-revenue-classification"] = map[string]any{"data": nil}
		} else {
			relationships["project-revenue-classification"] = map[string]any{
				"data": map[string]any{
					"type": "project-revenue-classifications",
					"id":   opts.ProjectRevenueClassification,
				},
			}
		}
		hasChanges = true
	}

	if cmd.Flags().Changed("quantity-estimate") {
		if opts.QuantityEstimate == "" {
			relationships["quantity-estimate"] = map[string]any{"data": nil}
		} else {
			relationships["quantity-estimate"] = map[string]any{
				"data": map[string]any{
					"type": "project-phase-revenue-item-quantity-estimates",
					"id":   opts.QuantityEstimate,
				},
			}
		}
		hasChanges = true
	}

	if !hasChanges {
		err := fmt.Errorf("no fields to update; specify at least one flag")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "project-phase-revenue-items",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/project-phase-revenue-items/"+opts.ID, jsonBody)
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

	row := buildProjectPhaseRevenueItemRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project phase revenue item %s\n", row.ID)
	return nil
}

func parseDoProjectPhaseRevenueItemsUpdateOptions(cmd *cobra.Command, args []string) (doProjectPhaseRevenueItemsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectRevenueClassification, _ := cmd.Flags().GetString("project-revenue-classification")
	quantityStrategy, _ := cmd.Flags().GetString("quantity-strategy")
	note, _ := cmd.Flags().GetString("note")
	quantityEstimate, _ := cmd.Flags().GetString("quantity-estimate")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectPhaseRevenueItemsUpdateOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		ID:                           args[0],
		ProjectRevenueClassification: projectRevenueClassification,
		QuantityStrategy:             quantityStrategy,
		Note:                         note,
		QuantityEstimate:             quantityEstimate,
	}, nil
}
