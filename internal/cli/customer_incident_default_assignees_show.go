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

type customerIncidentDefaultAssigneesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type customerIncidentDefaultAssigneeDetails struct {
	ID                string `json:"id"`
	CustomerID        string `json:"customer_id,omitempty"`
	Customer          string `json:"customer,omitempty"`
	DefaultAssigneeID string `json:"default_assignee_id,omitempty"`
	DefaultAssignee   string `json:"default_assignee,omitempty"`
	Kind              string `json:"kind,omitempty"`
	CreatedAt         string `json:"created_at,omitempty"`
	UpdatedAt         string `json:"updated_at,omitempty"`
}

func newCustomerIncidentDefaultAssigneesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show customer incident default assignee details",
		Long: `Show the full details of a customer incident default assignee.

Output Fields:
  ID
  Customer
  Default Assignee
  Kind
  Created At
  Updated At

Arguments:
  <id>    The default assignee ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a default assignee
  xbe view customer-incident-default-assignees show 123

  # Output as JSON
  xbe view customer-incident-default-assignees show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runCustomerIncidentDefaultAssigneesShow,
	}
	initCustomerIncidentDefaultAssigneesShowFlags(cmd)
	return cmd
}

func init() {
	customerIncidentDefaultAssigneesCmd.AddCommand(newCustomerIncidentDefaultAssigneesShowCmd())
}

func initCustomerIncidentDefaultAssigneesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCustomerIncidentDefaultAssigneesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseCustomerIncidentDefaultAssigneesShowOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("customer incident default assignee id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[customer-incident-default-assignees]", "kind,created-at,updated-at,customer,default-assignee")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[users]", "name,email-address")
	query.Set("include", "customer,default-assignee")

	body, _, err := client.Get(cmd.Context(), "/v1/customer-incident-default-assignees/"+id, query)
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

	details := buildCustomerIncidentDefaultAssigneeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderCustomerIncidentDefaultAssigneeDetails(cmd, details)
}

func parseCustomerIncidentDefaultAssigneesShowOptions(cmd *cobra.Command) (customerIncidentDefaultAssigneesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return customerIncidentDefaultAssigneesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildCustomerIncidentDefaultAssigneeDetails(resp jsonAPISingleResponse) customerIncidentDefaultAssigneeDetails {
	resource := resp.Data
	attrs := resource.Attributes
	included := map[string]map[string]any{}
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc.Attributes
	}

	details := customerIncidentDefaultAssigneeDetails{
		ID:        resource.ID,
		Kind:      stringAttr(attrs, "kind"),
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	details.CustomerID = relationshipIDFromMap(resource.Relationships, "customer")
	details.Customer = resolveCustomerIncidentDefaultAssigneeCustomerName(details.CustomerID, included)
	details.DefaultAssigneeID = relationshipIDFromMap(resource.Relationships, "default-assignee")
	details.DefaultAssignee = resolveCustomerIncidentDefaultAssigneeUserName(details.DefaultAssigneeID, included)

	return details
}

func resolveCustomerIncidentDefaultAssigneeCustomerName(id string, included map[string]map[string]any) string {
	if id == "" {
		return ""
	}
	if attrs, ok := included[resourceKey("customers", id)]; ok {
		return firstNonEmpty(stringAttr(attrs, "company-name"), stringAttr(attrs, "name"))
	}
	return ""
}

func resolveCustomerIncidentDefaultAssigneeUserName(id string, included map[string]map[string]any) string {
	if id == "" {
		return ""
	}
	if attrs, ok := included[resourceKey("users", id)]; ok {
		return firstNonEmpty(stringAttr(attrs, "full-name"), stringAttr(attrs, "name"), stringAttr(attrs, "email-address"))
	}
	return ""
}

func renderCustomerIncidentDefaultAssigneeDetails(cmd *cobra.Command, details customerIncidentDefaultAssigneeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Kind != "" {
		fmt.Fprintf(out, "Kind: %s\n", details.Kind)
	}
	if details.CustomerID != "" {
		label := details.CustomerID
		if details.Customer != "" {
			label = fmt.Sprintf("%s (%s)", details.Customer, details.CustomerID)
		}
		fmt.Fprintf(out, "Customer: %s\n", label)
	}
	if details.DefaultAssigneeID != "" {
		label := details.DefaultAssigneeID
		if details.DefaultAssignee != "" {
			label = fmt.Sprintf("%s (%s)", details.DefaultAssignee, details.DefaultAssigneeID)
		}
		fmt.Fprintf(out, "Default Assignee: %s\n", label)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
