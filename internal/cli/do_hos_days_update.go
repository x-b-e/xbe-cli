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

type doHosDaysUpdateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	ID                string
	RegulationSetCode string
}

func newDoHosDaysUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an HOS day",
		Long: `Update an existing HOS day.

Optional flags:
  --regulation-set-code  Regulation set code

Note: Only admin users can update HOS days.`,
		Example: `  # Update regulation set code
  xbe do hos-days update 123 --regulation-set-code US-80

  # Get JSON output
  xbe do hos-days update 123 --regulation-set-code US-80 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoHosDaysUpdate,
	}
	initDoHosDaysUpdateFlags(cmd)
	return cmd
}

func init() {
	doHosDaysCmd.AddCommand(newDoHosDaysUpdateCmd())
}

func initDoHosDaysUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("regulation-set-code", "", "Regulation set code")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoHosDaysUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoHosDaysUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("regulation-set-code") {
		attributes["regulation-set-code"] = opts.RegulationSetCode
	}

	if len(attributes) == 0 {
		err := errors.New("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":       "hos-days",
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

	path := fmt.Sprintf("/v1/hos-days/%s", opts.ID)
	body, _, err := client.Patch(cmd.Context(), path, jsonBody)
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

	row := buildHosDayRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated HOS day %s\n", row.ID)
	return nil
}

func parseDoHosDaysUpdateOptions(cmd *cobra.Command, args []string) (doHosDaysUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	regulationSetCode, _ := cmd.Flags().GetString("regulation-set-code")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doHosDaysUpdateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		ID:                args[0],
		RegulationSetCode: regulationSetCode,
	}, nil
}
