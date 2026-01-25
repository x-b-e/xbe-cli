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

type doTimeCardTimeChangesUpdateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	ID                     string
	TimeChangesAttributes  string
	Comment                string
	SkipQuantityValidation string
}

func newDoTimeCardTimeChangesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a time card time change",
		Long: `Update a time card time change.

Optional flags:
  --time-changes-attributes  Time change attributes as JSON object
  --comment                  Comment describing the change
  --skip-quantity-validation Skip quantity validation (true/false)

Note: Only unprocessed time card time changes can be updated.`,
		Example: `  # Update the comment
  xbe do time-card-time-changes update 123 --comment "Updated request"

  # Update time changes attributes
  xbe do time-card-time-changes update 123 --time-changes-attributes '{"down_minutes":30}'`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTimeCardTimeChangesUpdate,
	}
	initDoTimeCardTimeChangesUpdateFlags(cmd)
	return cmd
}

func init() {
	doTimeCardTimeChangesCmd.AddCommand(newDoTimeCardTimeChangesUpdateCmd())
}

func initDoTimeCardTimeChangesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("time-changes-attributes", "", "Time change attributes as JSON object")
	cmd.Flags().String("comment", "", "Comment describing the change")
	cmd.Flags().String("skip-quantity-validation", "", "Skip quantity validation (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTimeCardTimeChangesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTimeCardTimeChangesUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("time-changes-attributes") {
		if strings.TrimSpace(opts.TimeChangesAttributes) == "" {
			err := fmt.Errorf("--time-changes-attributes cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		var timeChanges map[string]any
		if err := json.Unmarshal([]byte(opts.TimeChangesAttributes), &timeChanges); err != nil {
			err := fmt.Errorf("invalid time-changes-attributes JSON: %w", err)
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		if len(timeChanges) == 0 {
			err := fmt.Errorf("--time-changes-attributes must include at least one change")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["time-changes-attributes"] = timeChanges
	}

	if cmd.Flags().Changed("comment") {
		attributes["comment"] = opts.Comment
	}

	setBoolAttrIfPresent(attributes, "skip-quantity-validation", opts.SkipQuantityValidation)

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "time-card-time-changes",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/time-card-time-changes/"+opts.ID, jsonBody)
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

	row := buildTimeCardTimeChangeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated time card time change %s\n", row.ID)
	return nil
}

func parseDoTimeCardTimeChangesUpdateOptions(cmd *cobra.Command, args []string) (doTimeCardTimeChangesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	timeChangesAttributes, _ := cmd.Flags().GetString("time-changes-attributes")
	comment, _ := cmd.Flags().GetString("comment")
	skipQuantityValidation, _ := cmd.Flags().GetString("skip-quantity-validation")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeCardTimeChangesUpdateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		ID:                     args[0],
		TimeChangesAttributes:  timeChangesAttributes,
		Comment:                comment,
		SkipQuantityValidation: skipQuantityValidation,
	}, nil
}
