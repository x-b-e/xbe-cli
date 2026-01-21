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

type doShiftFeedbackReasonsUpdateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	ID               string
	Name             string
	Kind             string
	Slug             string
	DefaultRating    string
	CorrectiveAction string
}

func newDoShiftFeedbackReasonsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing shift feedback reason",
		Long: `Update an existing shift feedback reason.

Provide the reason ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --name               The reason name
  --kind               The feedback kind
  --slug               URL-friendly identifier
  --default-rating     Default rating value
  --corrective-action  Corrective action text`,
		Example: `  # Update name
  xbe do shift-feedback-reasons update 123 --name "Updated Name"

  # Update multiple fields
  xbe do shift-feedback-reasons update 123 --name "New Name" --corrective-action "New action"

  # Get JSON output
  xbe do shift-feedback-reasons update 123 --name "New Name" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoShiftFeedbackReasonsUpdate,
	}
	initDoShiftFeedbackReasonsUpdateFlags(cmd)
	return cmd
}

func init() {
	doShiftFeedbackReasonsCmd.AddCommand(newDoShiftFeedbackReasonsUpdateCmd())
}

func initDoShiftFeedbackReasonsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Reason name")
	cmd.Flags().String("kind", "", "Feedback kind")
	cmd.Flags().String("slug", "", "URL-friendly identifier")
	cmd.Flags().String("default-rating", "", "Default rating value")
	cmd.Flags().String("corrective-action", "", "Corrective action text")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoShiftFeedbackReasonsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoShiftFeedbackReasonsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("kind") {
		attributes["kind"] = opts.Kind
	}
	if cmd.Flags().Changed("slug") {
		attributes["slug"] = opts.Slug
	}
	if cmd.Flags().Changed("default-rating") {
		attributes["default-rating"] = opts.DefaultRating
	}
	if cmd.Flags().Changed("corrective-action") {
		attributes["corrective-action"] = opts.CorrectiveAction
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one of --name, --kind, --slug, --default-rating, --corrective-action")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "shift-feedback-reasons",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/shift-feedback-reasons/"+opts.ID, jsonBody)
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

	row := buildShiftFeedbackReasonRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated shift feedback reason %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoShiftFeedbackReasonsUpdateOptions(cmd *cobra.Command, args []string) (doShiftFeedbackReasonsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	kind, _ := cmd.Flags().GetString("kind")
	slug, _ := cmd.Flags().GetString("slug")
	defaultRating, _ := cmd.Flags().GetString("default-rating")
	correctiveAction, _ := cmd.Flags().GetString("corrective-action")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doShiftFeedbackReasonsUpdateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		ID:               args[0],
		Name:             name,
		Kind:             kind,
		Slug:             slug,
		DefaultRating:    defaultRating,
		CorrectiveAction: correctiveAction,
	}, nil
}
