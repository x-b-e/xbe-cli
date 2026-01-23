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

type doDriverDayConstraintsUpdateOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	ID         string
	Constraint string
}

func newDoDriverDayConstraintsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a driver day constraint",
		Long: `Update a driver day constraint.

Note: driver-day cannot be changed after creation.

Optional flags:
  --constraint   Shift set time card constraint ID (use empty to clear)`,
		Example: `  # Update a driver day constraint
  xbe do driver-day-constraints update 123 --constraint 456

  # Clear a constraint
  xbe do driver-day-constraints update 123 --constraint ""`,
		Args: cobra.ExactArgs(1),
		RunE: runDoDriverDayConstraintsUpdate,
	}
	initDoDriverDayConstraintsUpdateFlags(cmd)
	return cmd
}

func init() {
	doDriverDayConstraintsCmd.AddCommand(newDoDriverDayConstraintsUpdateCmd())
}

func initDoDriverDayConstraintsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("constraint", "", "Shift set time card constraint ID (use empty to clear)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDriverDayConstraintsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoDriverDayConstraintsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("constraint") {
		if opts.Constraint == "" {
			relationships["constraint"] = map[string]any{"data": nil}
		} else {
			relationships["constraint"] = map[string]any{
				"data": map[string]any{
					"type": "shift-set-time-card-constraints",
					"id":   opts.Constraint,
				},
			}
		}
	}

	if len(relationships) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":          "driver-day-constraints",
		"id":            opts.ID,
		"relationships": relationships,
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/driver-day-constraints/"+opts.ID, jsonBody)
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

	row := buildDriverDayConstraintRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated driver day constraint %s\n", row.ID)
	return nil
}

func parseDoDriverDayConstraintsUpdateOptions(cmd *cobra.Command, args []string) (doDriverDayConstraintsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	constraint, _ := cmd.Flags().GetString("constraint")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDriverDayConstraintsUpdateOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		ID:         args[0],
		Constraint: constraint,
	}, nil
}
