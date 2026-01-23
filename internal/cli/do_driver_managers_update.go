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

type doDriverManagersUpdateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	ID                string
	TruckerID         string
	ManagerMembership string
	ManagedMembership string
}

func newDoDriverManagersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a driver manager",
		Long: `Update a driver manager.

Updatable fields:
  --trucker             Trucker ID
  --manager-membership  Manager membership ID
  --managed-membership  Managed membership ID`,
		Example: `  # Update manager membership
  xbe do driver-managers update 123 --manager-membership 456

  # Update managed membership
  xbe do driver-managers update 123 --managed-membership 789`,
		Args: cobra.ExactArgs(1),
		RunE: runDoDriverManagersUpdate,
	}
	initDoDriverManagersUpdateFlags(cmd)
	return cmd
}

func init() {
	doDriverManagersCmd.AddCommand(newDoDriverManagersUpdateCmd())
}

func initDoDriverManagersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("trucker", "", "Trucker ID")
	cmd.Flags().String("manager-membership", "", "Manager membership ID")
	cmd.Flags().String("managed-membership", "", "Managed membership ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDriverManagersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoDriverManagersUpdateOptions(cmd, args)
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

	relationships := map[string]any{}

	if cmd.Flags().Changed("trucker") {
		if strings.TrimSpace(opts.TruckerID) == "" {
			err := fmt.Errorf("--trucker cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["trucker"] = map[string]any{
			"data": map[string]any{
				"type": "truckers",
				"id":   opts.TruckerID,
			},
		}
	}
	if cmd.Flags().Changed("manager-membership") {
		if strings.TrimSpace(opts.ManagerMembership) == "" {
			err := fmt.Errorf("--manager-membership cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["manager-membership"] = map[string]any{
			"data": map[string]any{
				"type": "trucker-memberships",
				"id":   opts.ManagerMembership,
			},
		}
	}
	if cmd.Flags().Changed("managed-membership") {
		if strings.TrimSpace(opts.ManagedMembership) == "" {
			err := fmt.Errorf("--managed-membership cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["managed-membership"] = map[string]any{
			"data": map[string]any{
				"type": "trucker-memberships",
				"id":   opts.ManagedMembership,
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
			"type":          "driver-managers",
			"id":            opts.ID,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/driver-managers/"+opts.ID, jsonBody)
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

	row := buildDriverManagerRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated driver manager %s\n", row.ID)
	return nil
}

func parseDoDriverManagersUpdateOptions(cmd *cobra.Command, args []string) (doDriverManagersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	truckerID, _ := cmd.Flags().GetString("trucker")
	managerMembership, _ := cmd.Flags().GetString("manager-membership")
	managedMembership, _ := cmd.Flags().GetString("managed-membership")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDriverManagersUpdateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		ID:                args[0],
		TruckerID:         truckerID,
		ManagerMembership: managerMembership,
		ManagedMembership: managedMembership,
	}, nil
}
