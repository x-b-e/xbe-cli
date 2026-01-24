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

type doProjectRevenueItemsCreateOptions struct {
	BaseURL                        string
	Token                          string
	JSON                           bool
	Project                        string
	RevenueClassification          string
	UnitOfMeasure                  string
	Description                    string
	ExternalDeveloperRevenueItemID string
	DeveloperQuantityEstimate      string
}

func newDoProjectRevenueItemsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project revenue item",
		Long: `Create a project revenue item.

Required flags:
  --project                Project ID (required)
  --revenue-classification Revenue classification ID (required)
  --unit-of-measure        Unit of measure ID (required)

Optional flags:
  --description                       Revenue item description
  --external-developer-revenue-item-id External developer revenue item ID
  --developer-quantity-estimate       Developer quantity estimate

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a project revenue item
  xbe do project-revenue-items create \
    --project 123 \
    --revenue-classification 456 \
    --unit-of-measure 789 \
    --description "Base material"

  # Create with external developer ID and estimate
  xbe do project-revenue-items create \
    --project 123 \
    --revenue-classification 456 \
    --unit-of-measure 789 \
    --external-developer-revenue-item-id "HB-001" \
    --developer-quantity-estimate 1200`,
		Args: cobra.NoArgs,
		RunE: runDoProjectRevenueItemsCreate,
	}
	initDoProjectRevenueItemsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectRevenueItemsCmd.AddCommand(newDoProjectRevenueItemsCreateCmd())
}

func initDoProjectRevenueItemsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project", "", "Project ID (required)")
	cmd.Flags().String("revenue-classification", "", "Revenue classification ID (required)")
	cmd.Flags().String("unit-of-measure", "", "Unit of measure ID (required)")
	cmd.Flags().String("description", "", "Revenue item description")
	cmd.Flags().String("external-developer-revenue-item-id", "", "External developer revenue item ID")
	cmd.Flags().String("developer-quantity-estimate", "", "Developer quantity estimate")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectRevenueItemsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectRevenueItemsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Project) == "" {
		err := fmt.Errorf("--project is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.RevenueClassification) == "" {
		err := fmt.Errorf("--revenue-classification is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.UnitOfMeasure) == "" {
		err := fmt.Errorf("--unit-of-measure is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.Description != "" {
		attributes["description"] = opts.Description
	}
	if opts.ExternalDeveloperRevenueItemID != "" {
		attributes["external-developer-revenue-item-id"] = opts.ExternalDeveloperRevenueItemID
	}
	if opts.DeveloperQuantityEstimate != "" {
		attributes["developer-quantity-estimate"] = opts.DeveloperQuantityEstimate
	}

	relationships := map[string]any{
		"project": map[string]any{
			"data": map[string]any{
				"type": "projects",
				"id":   opts.Project,
			},
		},
		"revenue-classification": map[string]any{
			"data": map[string]any{
				"type": "project-revenue-classifications",
				"id":   opts.RevenueClassification,
			},
		},
		"unit-of-measure": map[string]any{
			"data": map[string]any{
				"type": "unit-of-measures",
				"id":   opts.UnitOfMeasure,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-revenue-items",
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-revenue-items", jsonBody)
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

	row := buildProjectRevenueItemRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project revenue item %s\n", row.ID)
	return nil
}

func parseDoProjectRevenueItemsCreateOptions(cmd *cobra.Command) (doProjectRevenueItemsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	project, _ := cmd.Flags().GetString("project")
	revenueClassification, _ := cmd.Flags().GetString("revenue-classification")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	description, _ := cmd.Flags().GetString("description")
	externalDeveloperRevenueItemID, _ := cmd.Flags().GetString("external-developer-revenue-item-id")
	developerQuantityEstimate, _ := cmd.Flags().GetString("developer-quantity-estimate")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectRevenueItemsCreateOptions{
		BaseURL:                        baseURL,
		Token:                          token,
		JSON:                           jsonOut,
		Project:                        project,
		RevenueClassification:          revenueClassification,
		UnitOfMeasure:                  unitOfMeasure,
		Description:                    description,
		ExternalDeveloperRevenueItemID: externalDeveloperRevenueItemID,
		DeveloperQuantityEstimate:      developerQuantityEstimate,
	}, nil
}

func buildProjectRevenueItemRowFromSingle(resp jsonAPISingleResponse) projectRevenueItemRow {
	return buildProjectRevenueItemRow(resp.Data, nil)
}
