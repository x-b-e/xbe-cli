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

type doDriverDayConstraintsCreateOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	DriverDay  string
	Constraint string
}

func newDoDriverDayConstraintsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a driver day constraint",
		Long: `Create a driver day constraint.

Required flags:
  --driver-day   Driver day ID (required)

Optional flags:
  --constraint   Shift set time card constraint ID`,
		Example: `  # Create a driver day constraint
  xbe do driver-day-constraints create --driver-day 123

  # Create with a constraint
  xbe do driver-day-constraints create --driver-day 123 --constraint 456`,
		Args: cobra.NoArgs,
		RunE: runDoDriverDayConstraintsCreate,
	}
	initDoDriverDayConstraintsCreateFlags(cmd)
	return cmd
}

func init() {
	doDriverDayConstraintsCmd.AddCommand(newDoDriverDayConstraintsCreateCmd())
}

func initDoDriverDayConstraintsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("driver-day", "", "Driver day ID (required)")
	cmd.Flags().String("constraint", "", "Shift set time card constraint ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDriverDayConstraintsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoDriverDayConstraintsCreateOptions(cmd)
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

	if opts.DriverDay == "" {
		err := fmt.Errorf("--driver-day is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"driver-day": map[string]any{
			"data": map[string]any{
				"type": "trucker-shift-sets",
				"id":   opts.DriverDay,
			},
		},
	}

	if opts.Constraint != "" {
		relationships["constraint"] = map[string]any{
			"data": map[string]any{
				"type": "shift-set-time-card-constraints",
				"id":   opts.Constraint,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "driver-day-constraints",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/driver-day-constraints", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created driver day constraint %s\n", row.ID)
	return nil
}

func parseDoDriverDayConstraintsCreateOptions(cmd *cobra.Command) (doDriverDayConstraintsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	driverDay, _ := cmd.Flags().GetString("driver-day")
	constraint, _ := cmd.Flags().GetString("constraint")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDriverDayConstraintsCreateOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		DriverDay:  driverDay,
		Constraint: constraint,
	}, nil
}
