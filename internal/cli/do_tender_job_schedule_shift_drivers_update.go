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

type doTenderJobScheduleShiftDriversUpdateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	ID        string
	IsPrimary bool
}

func newDoTenderJobScheduleShiftDriversUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a tender job schedule shift driver",
		Long: `Update a tender job schedule shift driver.

Optional flags:
  --is-primary  Mark the driver as primary (true/false)`,
		Example: `  # Mark a shift driver as primary
  xbe do tender-job-schedule-shift-drivers update 123 --is-primary true`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTenderJobScheduleShiftDriversUpdate,
	}
	initDoTenderJobScheduleShiftDriversUpdateFlags(cmd)
	return cmd
}

func init() {
	doTenderJobScheduleShiftDriversCmd.AddCommand(newDoTenderJobScheduleShiftDriversUpdateCmd())
}

func initDoTenderJobScheduleShiftDriversUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("is-primary", false, "Mark the driver as primary")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTenderJobScheduleShiftDriversUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTenderJobScheduleShiftDriversUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("is-primary") {
		attributes["is-primary"] = opts.IsPrimary
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "tender-job-schedule-shift-drivers",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/tender-job-schedule-shift-drivers/"+opts.ID, jsonBody)
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

	row := buildTenderJobScheduleShiftDriverRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated tender job schedule shift driver %s\n", row.ID)
	return nil
}

func parseDoTenderJobScheduleShiftDriversUpdateOptions(cmd *cobra.Command, args []string) (doTenderJobScheduleShiftDriversUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	isPrimary, _ := cmd.Flags().GetBool("is-primary")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTenderJobScheduleShiftDriversUpdateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		ID:        args[0],
		IsPrimary: isPrimary,
	}, nil
}
