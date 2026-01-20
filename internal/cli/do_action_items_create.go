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

type doActionItemsCreateOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	Title                   string
	Description             string
	DueOn                   string
	Status                  string
	Kind                    string
	ExpectedCostAmount      string
	ExpectedBenefitAmount   string
	RequiresXBEFeature      string
	ResponsibleOrganization string
}

func newDoActionItemsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new action item",
		Long: `Create a new action item.

Required flags:
  --title    The action item title (required)

Optional flags:
  --description                 Description text
  --due-on                      Due date (ISO 8601, e.g. 2024-12-31)
  --status                      Status: editing, ready_for_work, in_progress, in_verification, complete, on_hold
  --kind                        Kind: feature, integration, sombrero, bug_fix, change_management, data_seeding, training
  --expected-cost-amount        Expected cost amount
  --expected-benefit-amount     Expected benefit amount
  --requires-xbe-feature        Requires XBE feature (true/false)
  --responsible-organization    Responsible organization in Type|ID format (e.g. Broker|123)`,
		Example: `  # Create a simple action item
  xbe do action-items create --title "Fix production bug"

  # Create with status and kind
  xbe do action-items create --title "Add new feature" --status ready_for_work --kind feature

  # Create with due date and description
  xbe do action-items create --title "Q4 deliverable" --due-on 2024-12-31 --description "Deliver by end of Q4"

  # Create with responsible organization
  xbe do action-items create --title "Integration task" --responsible-organization Broker|123

  # Get JSON output
  xbe do action-items create --title "New task" --json`,
		Args: cobra.NoArgs,
		RunE: runDoActionItemsCreate,
	}
	initDoActionItemsCreateFlags(cmd)
	return cmd
}

func init() {
	doActionItemsCmd.AddCommand(newDoActionItemsCreateCmd())
}

func initDoActionItemsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("title", "", "Action item title (required)")
	cmd.Flags().String("description", "", "Description text")
	cmd.Flags().String("due-on", "", "Due date (ISO 8601, e.g. 2024-12-31)")
	cmd.Flags().String("status", "", "Status: editing, ready_for_work, in_progress, in_verification, complete, on_hold")
	cmd.Flags().String("kind", "", "Kind: feature, integration, sombrero, bug_fix, change_management, data_seeding, training")
	cmd.Flags().String("expected-cost-amount", "", "Expected cost amount")
	cmd.Flags().String("expected-benefit-amount", "", "Expected benefit amount")
	cmd.Flags().String("requires-xbe-feature", "", "Requires XBE feature (true/false)")
	cmd.Flags().String("responsible-organization", "", "Responsible organization in Type|ID format (e.g. Broker|123)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoActionItemsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoActionItemsCreateOptions(cmd)
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

	// Require title
	if opts.Title == "" {
		err := fmt.Errorf("--title is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Build attributes
	attributes := map[string]any{
		"title": opts.Title,
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
	}
	if opts.DueOn != "" {
		attributes["due-on"] = opts.DueOn
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

	// Build request data
	data := map[string]any{
		"type":       "action-items",
		"attributes": attributes,
	}

	// Build relationships if responsible-organization is provided
	if opts.ResponsibleOrganization != "" {
		orgType, orgID, err := parseOrganization(opts.ResponsibleOrganization)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		data["relationships"] = map[string]any{
			"responsible-organization": map[string]any{
				"data": map[string]string{
					"type": orgType,
					"id":   orgID,
				},
			},
		}
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

	body, _, err := client.Post(cmd.Context(), "/v1/action-items", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created action item %s (%s)\n", details.ID, details.Title)
	return renderActionItemDetails(cmd, details, actionItemsShowOptions{})
}

func parseDoActionItemsCreateOptions(cmd *cobra.Command) (doActionItemsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	title, _ := cmd.Flags().GetString("title")
	description, _ := cmd.Flags().GetString("description")
	dueOn, _ := cmd.Flags().GetString("due-on")
	status, _ := cmd.Flags().GetString("status")
	kind, _ := cmd.Flags().GetString("kind")
	expectedCostAmount, _ := cmd.Flags().GetString("expected-cost-amount")
	expectedBenefitAmount, _ := cmd.Flags().GetString("expected-benefit-amount")
	requiresXBEFeature, _ := cmd.Flags().GetString("requires-xbe-feature")
	responsibleOrganization, _ := cmd.Flags().GetString("responsible-organization")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doActionItemsCreateOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		Title:                   title,
		Description:             description,
		DueOn:                   dueOn,
		Status:                  status,
		Kind:                    kind,
		ExpectedCostAmount:      expectedCostAmount,
		ExpectedBenefitAmount:   expectedBenefitAmount,
		RequiresXBEFeature:      requiresXBEFeature,
		ResponsibleOrganization: responsibleOrganization,
	}, nil
}
