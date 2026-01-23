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

type doShiftFeedbacksUpdateOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	ID       string
	Rating   int
	Note     string
	ReasonID string
}

func newDoShiftFeedbacksUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a shift feedback",
		Long: `Update a shift feedback.

Optional flags:
  --rating                      Rating
  --note                        Feedback note
  --reason                      Shift feedback reason ID`,
		Example: `  # Update rating
  xbe do shift-feedbacks update 123 --rating 4

  # Update note
  xbe do shift-feedbacks update 123 --note "Updated feedback"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoShiftFeedbacksUpdate,
	}
	initDoShiftFeedbacksUpdateFlags(cmd)
	return cmd
}

func init() {
	doShiftFeedbacksCmd.AddCommand(newDoShiftFeedbacksUpdateCmd())
}

func initDoShiftFeedbacksUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Int("rating", 0, "Rating")
	cmd.Flags().String("note", "", "Feedback note")
	cmd.Flags().String("reason", "", "Shift feedback reason ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoShiftFeedbacksUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoShiftFeedbacksUpdateOptions(cmd, args)
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
	relationships := map[string]any{}

	if cmd.Flags().Changed("rating") {
		attributes["rating"] = opts.Rating
	}
	if cmd.Flags().Changed("note") {
		attributes["note"] = opts.Note
	}

	if cmd.Flags().Changed("reason") {
		if opts.ReasonID == "" {
			relationships["reason"] = map[string]any{"data": nil}
		} else {
			relationships["reason"] = map[string]any{
				"data": map[string]any{
					"type": "shift-feedback-reasons",
					"id":   opts.ReasonID,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "shift-feedbacks",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/shift-feedbacks/"+opts.ID, jsonBody)
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

	row := buildShiftFeedbackRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated shift feedback %s\n", row.ID)
	return nil
}

func parseDoShiftFeedbacksUpdateOptions(cmd *cobra.Command, args []string) (doShiftFeedbacksUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	rating, _ := cmd.Flags().GetInt("rating")
	note, _ := cmd.Flags().GetString("note")
	reasonID, _ := cmd.Flags().GetString("reason")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doShiftFeedbacksUpdateOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		ID:       args[0],
		Rating:   rating,
		Note:     note,
		ReasonID: reasonID,
	}, nil
}
