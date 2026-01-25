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

type doBusinessUnitMembershipsUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Kind    string
}

func newDoBusinessUnitMembershipsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a business unit membership",
		Long: `Update an existing business unit membership.

Arguments:
  <id>    The business unit membership ID (required)

Flags:
  --kind  Role (manager/technician/general)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update kind
  xbe do business-unit-memberships update 123 --kind manager

  # Get JSON output
  xbe do business-unit-memberships update 123 --kind technician --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoBusinessUnitMembershipsUpdate,
	}
	initDoBusinessUnitMembershipsUpdateFlags(cmd)
	return cmd
}

func init() {
	doBusinessUnitMembershipsCmd.AddCommand(newDoBusinessUnitMembershipsUpdateCmd())
}

func initDoBusinessUnitMembershipsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("kind", "", "Role (manager/technician/general)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBusinessUnitMembershipsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoBusinessUnitMembershipsUpdateOptions(cmd, args)
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

	attributes := map[string]any{}
	if cmd.Flags().Changed("kind") {
		if strings.TrimSpace(opts.Kind) == "" {
			err := fmt.Errorf("--kind cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["kind"] = opts.Kind
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify --kind")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":       "business-unit-memberships",
		"id":         opts.ID,
		"attributes": attributes,
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

	body, _, err := client.Patch(cmd.Context(), "/v1/business-unit-memberships/"+opts.ID, jsonBody)
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

	details := buildBusinessUnitMembershipDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated business unit membership %s\n", details.ID)
	return nil
}

func parseDoBusinessUnitMembershipsUpdateOptions(cmd *cobra.Command, args []string) (doBusinessUnitMembershipsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	kind, _ := cmd.Flags().GetString("kind")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBusinessUnitMembershipsUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Kind:    kind,
	}, nil
}
