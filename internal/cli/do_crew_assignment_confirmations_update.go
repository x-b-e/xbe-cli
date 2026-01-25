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

type doCrewAssignmentConfirmationsUpdateOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	ID         string
	Note       string
	IsExplicit bool
}

func newDoCrewAssignmentConfirmationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a crew assignment confirmation",
		Long: `Update a crew assignment confirmation.

Optional flags:
  --note         Update note
  --is-explicit  Update explicit flag`,
		Example: `  # Update note
  xbe do crew-assignment-confirmations update 123 --note "Confirmed"

  # Update explicit flag
  xbe do crew-assignment-confirmations update 123 --is-explicit true`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCrewAssignmentConfirmationsUpdate,
	}
	initDoCrewAssignmentConfirmationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doCrewAssignmentConfirmationsCmd.AddCommand(newDoCrewAssignmentConfirmationsUpdateCmd())
}

func initDoCrewAssignmentConfirmationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("note", "", "Confirmation note")
	cmd.Flags().Bool("is-explicit", false, "Mark confirmation as explicit")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCrewAssignmentConfirmationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCrewAssignmentConfirmationsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("note") {
		attributes["note"] = opts.Note
	}
	if cmd.Flags().Changed("is-explicit") {
		attributes["is-explicit"] = opts.IsExplicit
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "crew-assignment-confirmations",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/crew-assignment-confirmations/"+opts.ID, jsonBody)
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

	row := buildCrewAssignmentConfirmationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated crew assignment confirmation %s\n", row.ID)
	return nil
}

func parseDoCrewAssignmentConfirmationsUpdateOptions(cmd *cobra.Command, args []string) (doCrewAssignmentConfirmationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	note, _ := cmd.Flags().GetString("note")
	isExplicit, _ := cmd.Flags().GetBool("is-explicit")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCrewAssignmentConfirmationsUpdateOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		ID:         args[0],
		Note:       note,
		IsExplicit: isExplicit,
	}, nil
}
