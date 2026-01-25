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

type projectSubscriptionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectSubscriptionDetails struct {
	ID                      string `json:"id"`
	ProjectID               string `json:"project_id,omitempty"`
	ProjectName             string `json:"project_name,omitempty"`
	UserID                  string `json:"user_id,omitempty"`
	UserName                string `json:"user_name,omitempty"`
	UserEmail               string `json:"user_email,omitempty"`
	ContactMethod           string `json:"contact_method,omitempty"`
	CalculatedContactMethod string `json:"calculated_contact_method,omitempty"`
}

func newProjectSubscriptionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project subscription details",
		Long: `Show the full details of a project subscription.

Includes the associated project and user information.

Arguments:
  <id>  The project subscription ID (required).`,
		Example: `  # Show a subscription
  xbe view project-subscriptions show 123

  # Output as JSON
  xbe view project-subscriptions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectSubscriptionsShow,
	}
	initProjectSubscriptionsShowFlags(cmd)
	return cmd
}

func init() {
	projectSubscriptionsCmd.AddCommand(newProjectSubscriptionsShowCmd())
}

func initProjectSubscriptionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectSubscriptionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectSubscriptionsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("project subscription id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-subscriptions]", "project,user,contact-method,calculated-contact-method")
	query.Set("include", "project,user")
	query.Set("fields[projects]", "name")
	query.Set("fields[users]", "name,email-address")

	body, _, err := client.Get(cmd.Context(), "/v1/project-subscriptions/"+id, query)
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

	details := buildProjectSubscriptionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectSubscriptionDetails(cmd, details)
}

func parseProjectSubscriptionsShowOptions(cmd *cobra.Command) (projectSubscriptionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectSubscriptionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectSubscriptionDetails(resp jsonAPISingleResponse) projectSubscriptionDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := projectSubscriptionDetails{
		ID: resp.Data.ID,
	}

	details.ContactMethod = stringAttr(resp.Data.Attributes, "contact-method")
	details.CalculatedContactMethod = stringAttr(resp.Data.Attributes, "calculated-contact-method")

	if rel, ok := resp.Data.Relationships["project"]; ok && rel.Data != nil {
		details.ProjectID = rel.Data.ID
		if project, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ProjectName = stringAttr(project.Attributes, "name")
		}
	}

	if rel, ok := resp.Data.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.UserName = stringAttr(user.Attributes, "name")
			details.UserEmail = stringAttr(user.Attributes, "email-address")
		}
	}

	return details
}

func renderProjectSubscriptionDetails(cmd *cobra.Command, details projectSubscriptionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ProjectID != "" {
		fmt.Fprintf(out, "Project ID: %s\n", details.ProjectID)
	}
	if details.ProjectName != "" {
		fmt.Fprintf(out, "Project Name: %s\n", details.ProjectName)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User ID: %s\n", details.UserID)
	}
	if details.UserName != "" {
		fmt.Fprintf(out, "User Name: %s\n", details.UserName)
	}
	if details.UserEmail != "" {
		fmt.Fprintf(out, "User Email: %s\n", details.UserEmail)
	}
	if details.ContactMethod != "" {
		fmt.Fprintf(out, "Contact Method: %s\n", details.ContactMethod)
	}
	if details.CalculatedContactMethod != "" {
		fmt.Fprintf(out, "Calculated Contact Method: %s\n", details.CalculatedContactMethod)
	}

	return nil
}
