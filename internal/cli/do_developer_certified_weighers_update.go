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

type doDeveloperCertifiedWeighersUpdateOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	ID       string
	Number   string
	IsActive string
}

func newDoDeveloperCertifiedWeighersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a developer certified weigher",
		Long: `Update a developer certified weigher.

Only the fields you specify will be updated. Developer and user cannot be changed after creation.

Arguments:
  <id>    The developer certified weigher ID (required)

Optional flags:
  --number      Certified weigher number
  --is-active   Active status (true/false)`,
		Example: `  # Update a developer certified weigher number
  xbe do developer-certified-weighers update 123 --number CW-002

  # Deactivate a developer certified weigher
  xbe do developer-certified-weighers update 123 --is-active false

  # Update multiple fields
  xbe do developer-certified-weighers update 123 --number CW-003 --is-active true

  # Output as JSON
  xbe do developer-certified-weighers update 123 --number CW-003 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoDeveloperCertifiedWeighersUpdate,
	}
	initDoDeveloperCertifiedWeighersUpdateFlags(cmd)
	return cmd
}

func init() {
	doDeveloperCertifiedWeighersCmd.AddCommand(newDoDeveloperCertifiedWeighersUpdateCmd())
}

func initDoDeveloperCertifiedWeighersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("number", "", "Certified weigher number")
	cmd.Flags().String("is-active", "", "Active status (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDeveloperCertifiedWeighersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoDeveloperCertifiedWeighersUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("number") {
		attributes["number"] = opts.Number
	}
	if cmd.Flags().Changed("is-active") {
		attributes["is-active"] = opts.IsActive == "true"
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "developer-certified-weighers",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/developer-certified-weighers/"+opts.ID, jsonBody)
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

	row := developerCertifiedWeigherRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated developer certified weigher %s\n", row.ID)
	return nil
}

func parseDoDeveloperCertifiedWeighersUpdateOptions(cmd *cobra.Command, args []string) (doDeveloperCertifiedWeighersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	number, _ := cmd.Flags().GetString("number")
	isActive, _ := cmd.Flags().GetString("is-active")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	id := strings.TrimSpace(args[0])
	if id == "" {
		return doDeveloperCertifiedWeighersUpdateOptions{}, fmt.Errorf("developer certified weigher id is required")
	}

	return doDeveloperCertifiedWeighersUpdateOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		ID:       id,
		Number:   number,
		IsActive: isActive,
	}, nil
}
