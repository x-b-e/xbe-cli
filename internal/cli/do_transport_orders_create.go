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

type doTransportOrdersCreateOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	OrderedAt       string
	BillableMiles   string
	IsManaged       bool
	Status          string
	Customer        string
	Project         string
	ProjectDivision string
	ProjectOffice   string
	ProjectCategory string
}

func newDoTransportOrdersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new transport order",
		Long: `Create a new transport order.

Required flags:
  --customer          Customer ID

Optional flags:
  --ordered-at        Order datetime (ISO 8601)
  --billable-miles    Billable miles
  --is-managed        Mark as managed
  --status            Status

Relationships:
  --project           Project ID
  --project-division  Project division ID
  --project-office    Project office ID
  --project-category  Project category ID`,
		Example: `  # Create a transport order
  xbe do transport-orders create --customer 123

  # Create with project details
  xbe do transport-orders create --customer 123 --project 456 --ordered-at "2025-01-20T08:00:00Z"`,
		RunE: runDoTransportOrdersCreate,
	}
	initDoTransportOrdersCreateFlags(cmd)
	return cmd
}

func init() {
	doTransportOrdersCmd.AddCommand(newDoTransportOrdersCreateCmd())
}

func initDoTransportOrdersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("ordered-at", "", "Order datetime (ISO 8601)")
	cmd.Flags().String("billable-miles", "", "Billable miles")
	cmd.Flags().Bool("is-managed", false, "Mark as managed")
	cmd.Flags().String("status", "", "Status")
	cmd.Flags().String("customer", "", "Customer ID (required)")
	cmd.Flags().String("project", "", "Project ID")
	cmd.Flags().String("project-division", "", "Project division ID")
	cmd.Flags().String("project-office", "", "Project office ID")
	cmd.Flags().String("project-category", "", "Project category ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("customer")
}

func runDoTransportOrdersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTransportOrdersCreateOptions(cmd)
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

	if opts.OrderedAt != "" {
		attributes["ordered-at"] = opts.OrderedAt
	}
	if opts.BillableMiles != "" {
		attributes["billable-miles"] = opts.BillableMiles
	}
	if cmd.Flags().Changed("is-managed") {
		attributes["is-managed"] = opts.IsManaged
	}
	if opts.Status != "" {
		attributes["status"] = opts.Status
	}

	relationships := map[string]any{
		"customer": map[string]any{
			"data": map[string]any{
				"type": "customers",
				"id":   opts.Customer,
			},
		},
	}

	if opts.Project != "" {
		relationships["project"] = map[string]any{
			"data": map[string]any{
				"type": "projects",
				"id":   opts.Project,
			},
		}
	}
	if opts.ProjectDivision != "" {
		relationships["project-division"] = map[string]any{
			"data": map[string]any{
				"type": "project-divisions",
				"id":   opts.ProjectDivision,
			},
		}
	}
	if opts.ProjectOffice != "" {
		relationships["project-office"] = map[string]any{
			"data": map[string]any{
				"type": "project-offices",
				"id":   opts.ProjectOffice,
			},
		}
	}
	if opts.ProjectCategory != "" {
		relationships["project-category"] = map[string]any{
			"data": map[string]any{
				"type": "project-categories",
				"id":   opts.ProjectCategory,
			},
		}
	}

	data := map[string]any{
		"type":          "transport-orders",
		"attributes":    attributes,
		"relationships": relationships,
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

	body, _, err := client.Post(cmd.Context(), "/v1/transport-orders", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created transport order %s\n", resp.Data.ID)
	return nil
}

func parseDoTransportOrdersCreateOptions(cmd *cobra.Command) (doTransportOrdersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	orderedAt, _ := cmd.Flags().GetString("ordered-at")
	billableMiles, _ := cmd.Flags().GetString("billable-miles")
	isManaged, _ := cmd.Flags().GetBool("is-managed")
	status, _ := cmd.Flags().GetString("status")
	customer, _ := cmd.Flags().GetString("customer")
	project, _ := cmd.Flags().GetString("project")
	projectDivision, _ := cmd.Flags().GetString("project-division")
	projectOffice, _ := cmd.Flags().GetString("project-office")
	projectCategory, _ := cmd.Flags().GetString("project-category")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTransportOrdersCreateOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		OrderedAt:       orderedAt,
		BillableMiles:   billableMiles,
		IsManaged:       isManaged,
		Status:          status,
		Customer:        customer,
		Project:         project,
		ProjectDivision: projectDivision,
		ProjectOffice:   projectOffice,
		ProjectCategory: projectCategory,
	}, nil
}
