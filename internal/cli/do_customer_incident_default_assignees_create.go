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

type doCustomerIncidentDefaultAssigneesCreateOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	Customer        string
	DefaultAssignee string
	Kind            string
}

func newDoCustomerIncidentDefaultAssigneesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a customer incident default assignee",
		Long: `Create a customer incident default assignee.

Required flags:
  --customer          Customer ID (required)
  --default-assignee  User ID for the default assignee (required)
  --kind              Incident kind (required)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a default assignee
  xbe do customer-incident-default-assignees create --customer 123 --default-assignee 456 --kind safety

  # Output as JSON
  xbe do customer-incident-default-assignees create --customer 123 --default-assignee 456 --kind safety --json`,
		Args: cobra.NoArgs,
		RunE: runDoCustomerIncidentDefaultAssigneesCreate,
	}
	initDoCustomerIncidentDefaultAssigneesCreateFlags(cmd)
	return cmd
}

func init() {
	doCustomerIncidentDefaultAssigneesCmd.AddCommand(newDoCustomerIncidentDefaultAssigneesCreateCmd())
}

func initDoCustomerIncidentDefaultAssigneesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("customer", "", "Customer ID (required)")
	cmd.Flags().String("default-assignee", "", "User ID for the default assignee (required)")
	cmd.Flags().String("kind", "", "Incident kind (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCustomerIncidentDefaultAssigneesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoCustomerIncidentDefaultAssigneesCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	if opts.Customer == "" {
		err := fmt.Errorf("--customer is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.DefaultAssignee == "" {
		err := fmt.Errorf("--default-assignee is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Kind == "" {
		err := fmt.Errorf("--kind is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"kind": opts.Kind,
	}

	relationships := map[string]any{
		"customer": map[string]any{
			"data": map[string]any{
				"type": "customers",
				"id":   opts.Customer,
			},
		},
		"default-assignee": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.DefaultAssignee,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "customer-incident-default-assignees",
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

	body, _, err := client.Post(cmd.Context(), "/v1/customer-incident-default-assignees", jsonBody)
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

	row := customerIncidentDefaultAssigneeRowFromResource(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created customer incident default assignee %s\n", row.ID)
	return nil
}

func parseDoCustomerIncidentDefaultAssigneesCreateOptions(cmd *cobra.Command) (doCustomerIncidentDefaultAssigneesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	customer, _ := cmd.Flags().GetString("customer")
	defaultAssignee, _ := cmd.Flags().GetString("default-assignee")
	kind, _ := cmd.Flags().GetString("kind")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCustomerIncidentDefaultAssigneesCreateOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		Customer:        customer,
		DefaultAssignee: defaultAssignee,
		Kind:            kind,
	}, nil
}
