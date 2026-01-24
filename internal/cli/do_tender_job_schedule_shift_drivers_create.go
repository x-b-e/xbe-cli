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

type doTenderJobScheduleShiftDriversCreateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	TenderJobScheduleShift string
	User                   string
	IsPrimary              bool
}

func newDoTenderJobScheduleShiftDriversCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a tender job schedule shift driver",
		Long: `Create a tender job schedule shift driver.

Required flags:
  --tender-job-schedule-shift   Tender job schedule shift ID (required)
  --user                        User ID to assign as a driver (required)

Optional flags:
  --is-primary                   Mark the driver as primary`,
		Example: `  # Assign a driver to a shift
  xbe do tender-job-schedule-shift-drivers create \
    --tender-job-schedule-shift 123 \
    --user 456

  # Assign a primary driver
  xbe do tender-job-schedule-shift-drivers create \
    --tender-job-schedule-shift 123 \
    --user 456 \
    --is-primary`,
		Args: cobra.NoArgs,
		RunE: runDoTenderJobScheduleShiftDriversCreate,
	}
	initDoTenderJobScheduleShiftDriversCreateFlags(cmd)
	return cmd
}

func init() {
	doTenderJobScheduleShiftDriversCmd.AddCommand(newDoTenderJobScheduleShiftDriversCreateCmd())
}

func initDoTenderJobScheduleShiftDriversCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID (required)")
	cmd.Flags().String("user", "", "User ID to assign as a driver (required)")
	cmd.Flags().Bool("is-primary", false, "Mark the driver as primary")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTenderJobScheduleShiftDriversCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTenderJobScheduleShiftDriversCreateOptions(cmd)
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

	if opts.TenderJobScheduleShift == "" {
		err := fmt.Errorf("--tender-job-schedule-shift is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.User == "" {
		err := fmt.Errorf("--user is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("is-primary") {
		attributes["is-primary"] = opts.IsPrimary
	}

	relationships := map[string]any{
		"tender-job-schedule-shift": map[string]any{
			"data": map[string]any{
				"type": "tender-job-schedule-shifts",
				"id":   opts.TenderJobScheduleShift,
			},
		},
		"user": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.User,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "tender-job-schedule-shift-drivers",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/tender-job-schedule-shift-drivers", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created tender job schedule shift driver %s\n", row.ID)
	return nil
}

func parseDoTenderJobScheduleShiftDriversCreateOptions(cmd *cobra.Command) (doTenderJobScheduleShiftDriversCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	user, _ := cmd.Flags().GetString("user")
	isPrimary, _ := cmd.Flags().GetBool("is-primary")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTenderJobScheduleShiftDriversCreateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		TenderJobScheduleShift: tenderJobScheduleShift,
		User:                   user,
		IsPrimary:              isPrimary,
	}, nil
}
