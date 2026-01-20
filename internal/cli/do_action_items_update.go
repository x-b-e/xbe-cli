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

type doActionItemsUpdateOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	Title                 string
	Description           string
	DueOn                 string
	CompletedOn           string
	Status                string
	Kind                  string
	ExpectedCostAmount    string
	ExpectedBenefitAmount string
	RequiresXBEFeature    string
}

func newDoActionItemsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an action item",
		Long: `Update an existing action item.

Only the fields you specify will be updated. Fields not provided will remain unchanged.

Arguments:
  <id>    The action item ID (required)

Flags:
  --title                       Update the title
  --description                 Update the description
  --due-on                      Update the due date (ISO 8601, e.g. 2024-12-31)
  --completed-on                Set the completion date (ISO 8601)
  --status                      Update status: editing, ready_for_work, in_progress, in_verification, complete, on_hold
  --kind                        Update kind: feature, integration, sombrero, bug_fix, change_management, data_seeding, training
  --expected-cost-amount        Update expected cost amount
  --expected-benefit-amount     Update expected benefit amount
  --requires-xbe-feature        Update requires XBE feature (true/false)`,
		Example: `  # Update just the status
  xbe do action-items update 123 --status in_progress

  # Update multiple fields
  xbe do action-items update 123 --title "Updated title" --status complete --completed-on 2024-01-15

  # Update due date
  xbe do action-items update 123 --due-on 2024-12-31

  # Get JSON output
  xbe do action-items update 123 --status complete --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoActionItemsUpdate,
	}
	initDoActionItemsUpdateFlags(cmd)
	return cmd
}

func init() {
	doActionItemsCmd.AddCommand(newDoActionItemsUpdateCmd())
}

func initDoActionItemsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("title", "", "New title")
	cmd.Flags().String("description", "", "New description")
	cmd.Flags().String("due-on", "", "New due date (ISO 8601, e.g. 2024-12-31)")
	cmd.Flags().String("completed-on", "", "Completion date (ISO 8601)")
	cmd.Flags().String("status", "", "New status: editing, ready_for_work, in_progress, in_verification, complete, on_hold")
	cmd.Flags().String("kind", "", "New kind: feature, integration, sombrero, bug_fix, change_management, data_seeding, training")
	cmd.Flags().String("expected-cost-amount", "", "New expected cost amount")
	cmd.Flags().String("expected-benefit-amount", "", "New expected benefit amount")
	cmd.Flags().String("requires-xbe-feature", "", "Requires XBE feature (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoActionItemsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoActionItemsUpdateOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("action item id is required")
	}

	// Require at least one field to update
	if opts.Title == "" && opts.Description == "" && opts.DueOn == "" &&
		opts.CompletedOn == "" && opts.Status == "" && opts.Kind == "" &&
		opts.ExpectedCostAmount == "" && opts.ExpectedBenefitAmount == "" &&
		opts.RequiresXBEFeature == "" {
		err := fmt.Errorf("at least one field to update is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Build attributes
	attributes := map[string]any{}
	if opts.Title != "" {
		attributes["title"] = opts.Title
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
	}
	if opts.DueOn != "" {
		attributes["due-on"] = opts.DueOn
	}
	if opts.CompletedOn != "" {
		attributes["completed-on"] = opts.CompletedOn
	}
	if opts.Status != "" {
		attributes["status"] = opts.Status
	}
	if opts.Kind != "" {
		attributes["kind"] = opts.Kind
	}
	if opts.ExpectedCostAmount != "" {
		attributes["expected-cost-amount"] = opts.ExpectedCostAmount
	}
	if opts.ExpectedBenefitAmount != "" {
		attributes["expected-benefit-amount"] = opts.ExpectedBenefitAmount
	}
	if opts.RequiresXBEFeature != "" {
		attributes["requires-xbe-feature"] = opts.RequiresXBEFeature == "true"
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"id":         id,
			"type":       "action-items",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/action-items/"+id, jsonBody)
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

	details := buildActionItemDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderActionItemDetails(cmd, details, actionItemsShowOptions{})
}

func parseDoActionItemsUpdateOptions(cmd *cobra.Command) (doActionItemsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	title, _ := cmd.Flags().GetString("title")
	description, _ := cmd.Flags().GetString("description")
	dueOn, _ := cmd.Flags().GetString("due-on")
	completedOn, _ := cmd.Flags().GetString("completed-on")
	status, _ := cmd.Flags().GetString("status")
	kind, _ := cmd.Flags().GetString("kind")
	expectedCostAmount, _ := cmd.Flags().GetString("expected-cost-amount")
	expectedBenefitAmount, _ := cmd.Flags().GetString("expected-benefit-amount")
	requiresXBEFeature, _ := cmd.Flags().GetString("requires-xbe-feature")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doActionItemsUpdateOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		Title:                 title,
		Description:           description,
		DueOn:                 dueOn,
		CompletedOn:           completedOn,
		Status:                status,
		Kind:                  kind,
		ExpectedCostAmount:    expectedCostAmount,
		ExpectedBenefitAmount: expectedBenefitAmount,
		RequiresXBEFeature:    requiresXBEFeature,
	}, nil
}
