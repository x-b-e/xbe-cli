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

type doTransportOrdersUpdateOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	ID              string
	OrderedAt       string
	BillableMiles   string
	IsManaged       string
	Status          string
	Project         string
	ProjectDivision string
	ProjectOffice   string
	ProjectCategory string
}

func newDoTransportOrdersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a transport order",
		Long: `Update a transport order.

All flags are optional. Only provided flags will update the transport order.

Optional flags:
  --ordered-at        Order datetime (ISO 8601)
  --billable-miles    Billable miles
  --is-managed        Mark as managed (true/false)
  --status            Status

Relationships:
  --project           Project ID
  --project-division  Project division ID
  --project-office    Project office ID
  --project-category  Project category ID`,
		Example: `  # Update transport order status
  xbe do transport-orders update 123 --status "in_progress"

  # Update project association
  xbe do transport-orders update 123 --project 456

  # Mark as managed
  xbe do transport-orders update 123 --is-managed true`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTransportOrdersUpdate,
	}
	initDoTransportOrdersUpdateFlags(cmd)
	return cmd
}

func init() {
	doTransportOrdersCmd.AddCommand(newDoTransportOrdersUpdateCmd())
}

func initDoTransportOrdersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("ordered-at", "", "Order datetime (ISO 8601)")
	cmd.Flags().String("billable-miles", "", "Billable miles")
	cmd.Flags().String("is-managed", "", "Mark as managed (true/false)")
	cmd.Flags().String("status", "", "Status")
	cmd.Flags().String("project", "", "Project ID")
	cmd.Flags().String("project-division", "", "Project division ID")
	cmd.Flags().String("project-office", "", "Project office ID")
	cmd.Flags().String("project-category", "", "Project category ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTransportOrdersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTransportOrdersUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("ordered-at") {
		attributes["ordered-at"] = opts.OrderedAt
	}
	if cmd.Flags().Changed("billable-miles") {
		attributes["billable-miles"] = opts.BillableMiles
	}
	if cmd.Flags().Changed("is-managed") {
		attributes["is-managed"] = opts.IsManaged == "true"
	}
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}

	if cmd.Flags().Changed("project") {
		if opts.Project == "" {
			relationships["project"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["project"] = map[string]any{
				"data": map[string]any{
					"type": "projects",
					"id":   opts.Project,
				},
			}
		}
	}
	if cmd.Flags().Changed("project-division") {
		if opts.ProjectDivision == "" {
			relationships["project-division"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["project-division"] = map[string]any{
				"data": map[string]any{
					"type": "project-divisions",
					"id":   opts.ProjectDivision,
				},
			}
		}
	}
	if cmd.Flags().Changed("project-office") {
		if opts.ProjectOffice == "" {
			relationships["project-office"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["project-office"] = map[string]any{
				"data": map[string]any{
					"type": "project-offices",
					"id":   opts.ProjectOffice,
				},
			}
		}
	}
	if cmd.Flags().Changed("project-category") {
		if opts.ProjectCategory == "" {
			relationships["project-category"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["project-category"] = map[string]any{
				"data": map[string]any{
					"type": "project-categories",
					"id":   opts.ProjectCategory,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("at least one field must be specified for update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":       "transport-orders",
		"id":         opts.ID,
		"attributes": attributes,
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

	body, _, err := client.Patch(cmd.Context(), "/v1/transport-orders/"+opts.ID, jsonBody)
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
		return writeJSON(cmd.OutOrStdout(), map[string]string{
			"id":     resp.Data.ID,
			"status": stringAttr(resp.Data.Attributes, "status"),
		})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated transport order %s\n", resp.Data.ID)
	return nil
}

func parseDoTransportOrdersUpdateOptions(cmd *cobra.Command, args []string) (doTransportOrdersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	orderedAt, _ := cmd.Flags().GetString("ordered-at")
	billableMiles, _ := cmd.Flags().GetString("billable-miles")
	isManaged, _ := cmd.Flags().GetString("is-managed")
	status, _ := cmd.Flags().GetString("status")
	project, _ := cmd.Flags().GetString("project")
	projectDivision, _ := cmd.Flags().GetString("project-division")
	projectOffice, _ := cmd.Flags().GetString("project-office")
	projectCategory, _ := cmd.Flags().GetString("project-category")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTransportOrdersUpdateOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		ID:              args[0],
		OrderedAt:       orderedAt,
		BillableMiles:   billableMiles,
		IsManaged:       isManaged,
		Status:          status,
		Project:         project,
		ProjectDivision: projectDivision,
		ProjectOffice:   projectOffice,
		ProjectCategory: projectCategory,
	}, nil
}
