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

type doProjectSubscriptionsCreateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	Project       string
	User          string
	ContactMethod string
}

func newDoProjectSubscriptionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project subscription",
		Long: `Create a project subscription.

Required flags:
  --project  Project ID (required)
  --user     User ID (required)

Optional flags:
  --contact-method  Contact method (email_address, mobile_number)`,
		Example: `  # Subscribe a user to a project
  xbe do project-subscriptions create --project 123 --user 456

  # Set a contact method
  xbe do project-subscriptions create --project 123 --user 456 --contact-method email_address`,
		Args: cobra.NoArgs,
		RunE: runDoProjectSubscriptionsCreate,
	}
	initDoProjectSubscriptionsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectSubscriptionsCmd.AddCommand(newDoProjectSubscriptionsCreateCmd())
}

func initDoProjectSubscriptionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project", "", "Project ID (required)")
	cmd.Flags().String("user", "", "User ID (required)")
	cmd.Flags().String("contact-method", "", "Contact method (email_address, mobile_number)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectSubscriptionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectSubscriptionsCreateOptions(cmd)
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

	if opts.Project == "" {
		err := fmt.Errorf("--project is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.User == "" {
		err := fmt.Errorf("--user is required")
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
		"user": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.User,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-subscriptions",
			"relationships": relationships,
		},
	}

	if strings.TrimSpace(opts.ContactMethod) != "" {
		requestBody["data"].(map[string]any)["attributes"] = map[string]any{
			"contact-method": opts.ContactMethod,
		}
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/project-subscriptions", jsonBody)
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

	row := projectSubscriptionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project subscription %s\n", row.ID)
	return nil
}

func parseDoProjectSubscriptionsCreateOptions(cmd *cobra.Command) (doProjectSubscriptionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	project, _ := cmd.Flags().GetString("project")
	user, _ := cmd.Flags().GetString("user")
	contactMethod, _ := cmd.Flags().GetString("contact-method")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectSubscriptionsCreateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		Project:       project,
		User:          user,
		ContactMethod: contactMethod,
	}, nil
}
