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

type doMechanicUserAssociationsUpdateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	ID                     string
	User                   string
	MaintenanceRequirement string
}

func newDoMechanicUserAssociationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a mechanic user association",
		Long: `Update a mechanic user association.

Optional flags:
  --user                    User ID
  --maintenance-requirement Maintenance requirement ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update the user
  xbe do mechanic-user-associations update 123 --user 456

  # Update the maintenance requirement
  xbe do mechanic-user-associations update 123 --maintenance-requirement 789`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMechanicUserAssociationsUpdate,
	}
	initDoMechanicUserAssociationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doMechanicUserAssociationsCmd.AddCommand(newDoMechanicUserAssociationsUpdateCmd())
}

func initDoMechanicUserAssociationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("user", "", "User ID")
	cmd.Flags().String("maintenance-requirement", "", "Maintenance requirement ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMechanicUserAssociationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMechanicUserAssociationsUpdateOptions(cmd, args)
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

	id := strings.TrimSpace(opts.ID)
	if id == "" {
		return fmt.Errorf("mechanic user association id is required")
	}

	relationships := map[string]any{}
	if cmd.Flags().Changed("user") {
		if strings.TrimSpace(opts.User) == "" {
			err := fmt.Errorf("--user cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["user"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.User,
			},
		}
	}
	if cmd.Flags().Changed("maintenance-requirement") {
		if strings.TrimSpace(opts.MaintenanceRequirement) == "" {
			err := fmt.Errorf("--maintenance-requirement cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["maintenance-requirement"] = map[string]any{
			"data": map[string]any{
				"type": "maintenance-requirements",
				"id":   opts.MaintenanceRequirement,
			},
		}
	}

	if len(relationships) == 0 {
		err := fmt.Errorf("no relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "mechanic-user-associations",
			"id":            id,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/mechanic-user-associations/"+id, jsonBody)
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
		row := buildMechanicUserAssociationRowFromSingle(resp)
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated mechanic user association %s\n", resp.Data.ID)
	return nil
}

func parseDoMechanicUserAssociationsUpdateOptions(cmd *cobra.Command, args []string) (doMechanicUserAssociationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	user, _ := cmd.Flags().GetString("user")
	maintenanceRequirement, _ := cmd.Flags().GetString("maintenance-requirement")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMechanicUserAssociationsUpdateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		ID:                     args[0],
		User:                   user,
		MaintenanceRequirement: maintenanceRequirement,
	}, nil
}
