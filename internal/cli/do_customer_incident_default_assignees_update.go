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

type doCustomerIncidentDefaultAssigneesUpdateOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	ID              string
	DefaultAssignee string
	Kind            string
}

func newDoCustomerIncidentDefaultAssigneesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a customer incident default assignee",
		Long: `Update a customer incident default assignee.

Optional flags:
  --default-assignee  User ID for the default assignee
  --kind              Incident kind

Note: The customer relationship cannot be changed after creation.

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update the default assignee
  xbe do customer-incident-default-assignees update 123 --default-assignee 456

  # Update the kind
  xbe do customer-incident-default-assignees update 123 --kind quality`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCustomerIncidentDefaultAssigneesUpdate,
	}
	initDoCustomerIncidentDefaultAssigneesUpdateFlags(cmd)
	return cmd
}

func init() {
	doCustomerIncidentDefaultAssigneesCmd.AddCommand(newDoCustomerIncidentDefaultAssigneesUpdateCmd())
}

func initDoCustomerIncidentDefaultAssigneesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("default-assignee", "", "User ID for the default assignee")
	cmd.Flags().String("kind", "", "Incident kind")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCustomerIncidentDefaultAssigneesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCustomerIncidentDefaultAssigneesUpdateOptions(cmd, args)
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

	attributes := map[string]any{}
	relationships := map[string]any{}

	if cmd.Flags().Changed("kind") {
		if opts.Kind == "" {
			err := fmt.Errorf("--kind cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["kind"] = opts.Kind
	}

	if cmd.Flags().Changed("default-assignee") {
		if opts.DefaultAssignee == "" {
			err := fmt.Errorf("--default-assignee cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["default-assignee"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.DefaultAssignee,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":       "customer-incident-default-assignees",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/customer-incident-default-assignees/"+opts.ID, jsonBody)
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
		row := customerIncidentDefaultAssigneeRowFromResource(resp.Data)
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated customer incident default assignee %s\n", resp.Data.ID)
	return nil
}

func parseDoCustomerIncidentDefaultAssigneesUpdateOptions(cmd *cobra.Command, args []string) (doCustomerIncidentDefaultAssigneesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	defaultAssignee, _ := cmd.Flags().GetString("default-assignee")
	kind, _ := cmd.Flags().GetString("kind")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCustomerIncidentDefaultAssigneesUpdateOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		ID:              args[0],
		DefaultAssignee: defaultAssignee,
		Kind:            kind,
	}, nil
}
