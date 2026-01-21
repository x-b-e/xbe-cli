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

type doCraftsUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Name    string
	Code    string
}

func newDoCraftsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing craft",
		Long: `Update an existing craft.

Provide the craft ID as an argument, then use flags to specify which
fields to update. Only specified fields will be modified.

Updatable fields:
  --name    The craft name
  --code    Short code`,
		Example: `  # Update name
  xbe do crafts update 123 --name "Senior Carpenter"

  # Update code
  xbe do crafts update 123 --code "SR-CARP"

  # Update multiple fields
  xbe do crafts update 123 --name "Lead Electrician" --code "LEAD-ELEC"

  # Get JSON output
  xbe do crafts update 123 --name "Updated" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCraftsUpdate,
	}
	initDoCraftsUpdateFlags(cmd)
	return cmd
}

func init() {
	doCraftsCmd.AddCommand(newDoCraftsUpdateCmd())
}

func initDoCraftsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Craft name")
	cmd.Flags().String("code", "", "Short code")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCraftsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCraftsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
	}
	if cmd.Flags().Changed("code") {
		attributes["code"] = opts.Code
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one of --name, --code")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "crafts",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/crafts/"+opts.ID, jsonBody)
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

	row := buildCraftRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated craft %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoCraftsUpdateOptions(cmd *cobra.Command, args []string) (doCraftsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	code, _ := cmd.Flags().GetString("code")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCraftsUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Name:    name,
		Code:    code,
	}, nil
}
