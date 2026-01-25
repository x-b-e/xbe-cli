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

type doBusinessUnitMembershipsCreateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	BusinessUnit string
	Membership   string
	Kind         string
}

func newDoBusinessUnitMembershipsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a business unit membership",
		Long: `Create a business unit membership.

Required flags:
  --business-unit  Business unit ID (required)
  --membership     Membership ID (required)

Optional flags:
  --kind           Role (manager/technician/general)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a business unit membership
  xbe do business-unit-memberships create --business-unit 123 --membership 456

  # Create with a specific kind
  xbe do business-unit-memberships create --business-unit 123 --membership 456 --kind technician

  # Get JSON output
  xbe do business-unit-memberships create --business-unit 123 --membership 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoBusinessUnitMembershipsCreate,
	}
	initDoBusinessUnitMembershipsCreateFlags(cmd)
	return cmd
}

func init() {
	doBusinessUnitMembershipsCmd.AddCommand(newDoBusinessUnitMembershipsCreateCmd())
}

func initDoBusinessUnitMembershipsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("business-unit", "", "Business unit ID (required)")
	cmd.Flags().String("membership", "", "Membership ID (required)")
	cmd.Flags().String("kind", "", "Role (manager/technician/general)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBusinessUnitMembershipsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoBusinessUnitMembershipsCreateOptions(cmd)
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

	businessUnitID := strings.TrimSpace(opts.BusinessUnit)
	if businessUnitID == "" {
		err := fmt.Errorf("--business-unit is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	membershipID := strings.TrimSpace(opts.Membership)
	if membershipID == "" {
		err := fmt.Errorf("--membership is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if strings.TrimSpace(opts.Kind) != "" {
		attributes["kind"] = opts.Kind
	}

	relationships := map[string]any{
		"business-unit": map[string]any{
			"data": map[string]any{
				"type": "business-units",
				"id":   businessUnitID,
			},
		},
		"membership": map[string]any{
			"data": map[string]any{
				"type": "memberships",
				"id":   membershipID,
			},
		},
	}

	data := map[string]any{
		"type":          "business-unit-memberships",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
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

	body, _, err := client.Post(cmd.Context(), "/v1/business-unit-memberships", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created business unit membership %s\n", details.ID)
	return nil
}

func parseDoBusinessUnitMembershipsCreateOptions(cmd *cobra.Command) (doBusinessUnitMembershipsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	businessUnit, _ := cmd.Flags().GetString("business-unit")
	membership, _ := cmd.Flags().GetString("membership")
	kind, _ := cmd.Flags().GetString("kind")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBusinessUnitMembershipsCreateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		BusinessUnit: businessUnit,
		Membership:   membership,
		Kind:         kind,
	}, nil
}
