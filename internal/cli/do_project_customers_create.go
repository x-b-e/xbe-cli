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

type doProjectCustomersCreateOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	Project  string
	Customer string
}

func newDoProjectCustomersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project customer",
		Long: `Create a project customer.

Required flags:
  --project   Project ID
  --customer  Customer ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a project customer
  xbe do project-customers create --project 123 --customer 456`,
		Args: cobra.NoArgs,
		RunE: runDoProjectCustomersCreate,
	}
	initDoProjectCustomersCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectCustomersCmd.AddCommand(newDoProjectCustomersCreateCmd())
}

func initDoProjectCustomersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project", "", "Project ID")
	cmd.Flags().String("customer", "", "Customer ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectCustomersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectCustomersCreateOptions(cmd)
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
	if strings.TrimSpace(opts.Customer) == "" {
		err := fmt.Errorf("--customer is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"project": map[string]any{
			"data": map[string]any{
				"type": "projects",
				"id":   opts.Project,
			},
		},
		"customer": map[string]any{
			"data": map[string]any{
				"type": "customers",
				"id":   opts.Customer,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-customers",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/project-customers", jsonBody)
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
		row := buildProjectCustomerRows(jsonAPIResponse{Data: []jsonAPIResource{resp.Data}})
		if len(row) > 0 {
			return writeJSON(cmd.OutOrStdout(), row[0])
		}
		return writeJSON(cmd.OutOrStdout(), map[string]any{"id": resp.Data.ID})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project customer %s\n", resp.Data.ID)
	return nil
}

func parseDoProjectCustomersCreateOptions(cmd *cobra.Command) (doProjectCustomersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	project, _ := cmd.Flags().GetString("project")
	customer, _ := cmd.Flags().GetString("customer")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectCustomersCreateOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		Project:  project,
		Customer: customer,
	}, nil
}
