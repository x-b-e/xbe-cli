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

type doCostCodesUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	ID          string
	Description string
	IsActive    bool
}

func newDoCostCodesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a cost code",
		Long: `Update a cost code.

Optional flags:
  --description   Description of the cost code
  --active        Set as active
  --no-active     Set as inactive

Note: The code value and organization relationships cannot be changed after creation.`,
		Example: `  # Update cost code description
  xbe do cost-codes update 123 --description "Updated description"

  # Deactivate a cost code
  xbe do cost-codes update 123 --no-active`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCostCodesUpdate,
	}
	initDoCostCodesUpdateFlags(cmd)
	return cmd
}

func init() {
	doCostCodesCmd.AddCommand(newDoCostCodesUpdateCmd())
}

func initDoCostCodesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("description", "", "Description of the cost code")
	cmd.Flags().Bool("active", false, "Set as active")
	cmd.Flags().Bool("no-active", false, "Set as inactive")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCostCodesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCostCodesUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("active") {
		attributes["is-active"] = true
	}
	if cmd.Flags().Changed("no-active") {
		attributes["is-active"] = false
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "cost-codes",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/cost-codes/"+opts.ID, jsonBody)
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

	row := costCodeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated cost code %s\n", row.ID)
	return nil
}

func parseDoCostCodesUpdateOptions(cmd *cobra.Command, args []string) (doCostCodesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	description, _ := cmd.Flags().GetString("description")
	isActive, _ := cmd.Flags().GetBool("active")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCostCodesUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		ID:          args[0],
		Description: description,
		IsActive:    isActive,
	}, nil
}
