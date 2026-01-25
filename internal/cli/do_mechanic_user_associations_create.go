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

type doMechanicUserAssociationsCreateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	User                   string
	MaintenanceRequirement string
}

func newDoMechanicUserAssociationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a mechanic user association",
		Long: `Create a mechanic user association.

Required flags:
  --user                    User ID (required)
  --maintenance-requirement Maintenance requirement ID (required)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a mechanic user association
  xbe do mechanic-user-associations create --user 123 --maintenance-requirement 456

  # Output as JSON
  xbe do mechanic-user-associations create --user 123 --maintenance-requirement 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoMechanicUserAssociationsCreate,
	}
	initDoMechanicUserAssociationsCreateFlags(cmd)
	return cmd
}

func init() {
	doMechanicUserAssociationsCmd.AddCommand(newDoMechanicUserAssociationsCreateCmd())
}

func initDoMechanicUserAssociationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("user", "", "User ID (required)")
	cmd.Flags().String("maintenance-requirement", "", "Maintenance requirement ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMechanicUserAssociationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMechanicUserAssociationsCreateOptions(cmd)
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

	userID := strings.TrimSpace(opts.User)
	if userID == "" {
		err := fmt.Errorf("--user is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	maintenanceRequirementID := strings.TrimSpace(opts.MaintenanceRequirement)
	if maintenanceRequirementID == "" {
		err := fmt.Errorf("--maintenance-requirement is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"user": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   userID,
			},
		},
		"maintenance-requirement": map[string]any{
			"data": map[string]any{
				"type": "maintenance-requirements",
				"id":   maintenanceRequirementID,
			},
		},
	}

	data := map[string]any{
		"type":          "mechanic-user-associations",
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

	body, _, err := client.Post(cmd.Context(), "/v1/mechanic-user-associations", jsonBody)
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

	row := buildMechanicUserAssociationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created mechanic user association %s\n", row.ID)
	return nil
}

func parseDoMechanicUserAssociationsCreateOptions(cmd *cobra.Command) (doMechanicUserAssociationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	user, _ := cmd.Flags().GetString("user")
	maintenanceRequirement, _ := cmd.Flags().GetString("maintenance-requirement")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMechanicUserAssociationsCreateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		User:                   user,
		MaintenanceRequirement: maintenanceRequirement,
	}, nil
}

func buildMechanicUserAssociationRowFromSingle(resp jsonAPISingleResponse) mechanicUserAssociationRow {
	resource := resp.Data
	row := mechanicUserAssociationRow{ID: resource.ID}

	row.UserID = relationshipIDFromMap(resource.Relationships, "user")
	row.MaintenanceRequirementID = relationshipIDFromMap(resource.Relationships, "maintenance-requirement")

	return row
}
