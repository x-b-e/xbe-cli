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

type doProjectPhaseRevenueItemsCreateOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	ProjectPhase                 string
	ProjectRevenueItem           string
	ProjectRevenueClassification string
	QuantityStrategy             string
	Note                         string
}

func newDoProjectPhaseRevenueItemsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project phase revenue item",
		Long: `Create a project phase revenue item.

Required flags:
  --project-phase          Project phase ID
  --project-revenue-item   Project revenue item ID

Optional flags:
  --project-revenue-classification Project revenue classification ID
  --quantity-strategy              Quantity strategy (direct/indirect)
  --note                           Note

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a project phase revenue item
  xbe do project-phase-revenue-items create \
    --project-phase 123 \
    --project-revenue-item 456

  # Create with strategy and note
  xbe do project-phase-revenue-items create \
    --project-phase 123 \
    --project-revenue-item 456 \
    --quantity-strategy indirect \
    --note "Bid item"`,
		Args: cobra.NoArgs,
		RunE: runDoProjectPhaseRevenueItemsCreate,
	}
	initDoProjectPhaseRevenueItemsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectPhaseRevenueItemsCmd.AddCommand(newDoProjectPhaseRevenueItemsCreateCmd())
}

func initDoProjectPhaseRevenueItemsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-phase", "", "Project phase ID (required)")
	cmd.Flags().String("project-revenue-item", "", "Project revenue item ID (required)")
	cmd.Flags().String("project-revenue-classification", "", "Project revenue classification ID")
	cmd.Flags().String("quantity-strategy", "", "Quantity strategy (direct/indirect)")
	cmd.Flags().String("note", "", "Note")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("project-phase")
	_ = cmd.MarkFlagRequired("project-revenue-item")
}

func runDoProjectPhaseRevenueItemsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectPhaseRevenueItemsCreateOptions(cmd)
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

	projectPhase := strings.TrimSpace(opts.ProjectPhase)
	if projectPhase == "" {
		err := fmt.Errorf("--project-phase is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	projectRevenueItem := strings.TrimSpace(opts.ProjectRevenueItem)
	if projectRevenueItem == "" {
		err := fmt.Errorf("--project-revenue-item is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if strings.TrimSpace(opts.QuantityStrategy) != "" {
		attributes["quantity-strategy"] = strings.TrimSpace(opts.QuantityStrategy)
	}
	if cmd.Flags().Changed("note") {
		attributes["note"] = opts.Note
	}

	relationships := map[string]any{
		"project-phase": map[string]any{
			"data": map[string]any{
				"type": "project-phases",
				"id":   projectPhase,
			},
		},
		"project-revenue-item": map[string]any{
			"data": map[string]any{
				"type": "project-revenue-items",
				"id":   projectRevenueItem,
			},
		},
	}

	if strings.TrimSpace(opts.ProjectRevenueClassification) != "" {
		relationships["project-revenue-classification"] = map[string]any{
			"data": map[string]any{
				"type": "project-revenue-classifications",
				"id":   strings.TrimSpace(opts.ProjectRevenueClassification),
			},
		}
	}

	data := map[string]any{
		"type":          "project-phase-revenue-items",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-phase-revenue-items", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created project phase revenue item %s\n", row.ID)
	return nil
}

func parseDoProjectPhaseRevenueItemsCreateOptions(cmd *cobra.Command) (doProjectPhaseRevenueItemsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectPhase, _ := cmd.Flags().GetString("project-phase")
	projectRevenueItem, _ := cmd.Flags().GetString("project-revenue-item")
	projectRevenueClassification, _ := cmd.Flags().GetString("project-revenue-classification")
	quantityStrategy, _ := cmd.Flags().GetString("quantity-strategy")
	note, _ := cmd.Flags().GetString("note")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectPhaseRevenueItemsCreateOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		ProjectPhase:                 projectPhase,
		ProjectRevenueItem:           projectRevenueItem,
		ProjectRevenueClassification: projectRevenueClassification,
		QuantityStrategy:             quantityStrategy,
		Note:                         note,
	}, nil
}
