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

type doCraftClassesUpdateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	ID                string
	Name              string
	Code              string
	IsValidForDrivers bool
}

func newDoCraftClassesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing craft class",
		Long: `Update an existing craft class.

Provide the craft class ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --name                  The craft class name
  --code                  Craft class code
  --is-valid-for-drivers  Whether valid for drivers`,
		Example: `  # Update name
  xbe do craft-classes update 123 --name "Updated Name"

  # Update multiple fields
  xbe do craft-classes update 123 --name "New Name" --code "NEW"

  # Set valid for drivers
  xbe do craft-classes update 123 --is-valid-for-drivers`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCraftClassesUpdate,
	}
	initDoCraftClassesUpdateFlags(cmd)
	return cmd
}

func init() {
	doCraftClassesCmd.AddCommand(newDoCraftClassesUpdateCmd())
}

func initDoCraftClassesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Craft class name")
	cmd.Flags().String("code", "", "Craft class code")
	cmd.Flags().Bool("is-valid-for-drivers", false, "Whether valid for drivers")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCraftClassesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCraftClassesUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("is-valid-for-drivers") {
		attributes["is-valid-for-drivers"] = opts.IsValidForDrivers
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one of --name, --code, --is-valid-for-drivers")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "craft-classes",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/craft-classes/"+opts.ID, jsonBody)
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

	row := buildCraftClassRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated craft class %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoCraftClassesUpdateOptions(cmd *cobra.Command, args []string) (doCraftClassesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	code, _ := cmd.Flags().GetString("code")
	isValidForDrivers, _ := cmd.Flags().GetBool("is-valid-for-drivers")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCraftClassesUpdateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		ID:                args[0],
		Name:              name,
		Code:              code,
		IsValidForDrivers: isValidForDrivers,
	}, nil
}
