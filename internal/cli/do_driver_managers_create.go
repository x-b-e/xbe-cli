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

type doDriverManagersCreateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	TruckerID         string
	ManagerMembership string
	ManagedMembership string
}

func newDoDriverManagersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a driver manager",
		Long: `Create a driver manager.

Required flags:
  --trucker             Trucker ID (required)
  --manager-membership  Manager membership ID (required)
  --managed-membership  Managed membership ID (required)`,
		Example: `  # Create a driver manager
  xbe do driver-managers create \
    --trucker 123 \
    --manager-membership 456 \
    --managed-membership 789

  # Get JSON output
  xbe do driver-managers create \
    --trucker 123 \
    --manager-membership 456 \
    --managed-membership 789 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoDriverManagersCreate,
	}
	initDoDriverManagersCreateFlags(cmd)
	return cmd
}

func init() {
	doDriverManagersCmd.AddCommand(newDoDriverManagersCreateCmd())
}

func initDoDriverManagersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("trucker", "", "Trucker ID (required)")
	cmd.Flags().String("manager-membership", "", "Manager membership ID (required)")
	cmd.Flags().String("managed-membership", "", "Managed membership ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDriverManagersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoDriverManagersCreateOptions(cmd)
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

	if opts.TruckerID == "" {
		err := fmt.Errorf("--trucker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.ManagerMembership == "" {
		err := fmt.Errorf("--manager-membership is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.ManagedMembership == "" {
		err := fmt.Errorf("--managed-membership is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"trucker": map[string]any{
			"data": map[string]any{
				"type": "truckers",
				"id":   opts.TruckerID,
			},
		},
		"manager-membership": map[string]any{
			"data": map[string]any{
				"type": "trucker-memberships",
				"id":   opts.ManagerMembership,
			},
		},
		"managed-membership": map[string]any{
			"data": map[string]any{
				"type": "trucker-memberships",
				"id":   opts.ManagedMembership,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "driver-managers",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/driver-managers", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created driver manager %s\n", row.ID)
	return nil
}

func parseDoDriverManagersCreateOptions(cmd *cobra.Command) (doDriverManagersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	truckerID, _ := cmd.Flags().GetString("trucker")
	managerMembership, _ := cmd.Flags().GetString("manager-membership")
	managedMembership, _ := cmd.Flags().GetString("managed-membership")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDriverManagersCreateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		TruckerID:         truckerID,
		ManagerMembership: managerMembership,
		ManagedMembership: managedMembership,
	}, nil
}
