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

type doStakeholderClassificationsUpdateOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	ID                          string
	Title                       string
	LeverageFactor              int
	ObjectivesNarrativeExplicit string
}

func newDoStakeholderClassificationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing stakeholder classification",
		Long: `Update an existing stakeholder classification.

Provide the classification ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --title                          The classification title
  --leverage-factor                Influence level (integer)
  --objectives-narrative-explicit  Explicit objectives narrative

Note: Only admin users can update stakeholder classifications.
The slug is read-only and cannot be changed.`,
		Example: `  # Update title
  xbe do stakeholder-classifications update 123 --title "Primary Owner"

  # Update leverage factor
  xbe do stakeholder-classifications update 123 --leverage-factor 10

  # Update multiple fields
  xbe do stakeholder-classifications update 123 --title "Key Sponsor" --leverage-factor 8

  # Get JSON output
  xbe do stakeholder-classifications update 123 --title "Updated" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoStakeholderClassificationsUpdate,
	}
	initDoStakeholderClassificationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doStakeholderClassificationsCmd.AddCommand(newDoStakeholderClassificationsUpdateCmd())
}

func initDoStakeholderClassificationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("title", "", "Classification title")
	cmd.Flags().Int("leverage-factor", 0, "Leverage factor (influence level)")
	cmd.Flags().String("objectives-narrative-explicit", "", "Explicit objectives narrative")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoStakeholderClassificationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoStakeholderClassificationsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("title") {
		attributes["title"] = opts.Title
	}
	if cmd.Flags().Changed("leverage-factor") {
		attributes["leverage-factor"] = opts.LeverageFactor
	}
	if cmd.Flags().Changed("objectives-narrative-explicit") {
		attributes["objectives-narrative-explicit"] = opts.ObjectivesNarrativeExplicit
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one of --title, --leverage-factor, --objectives-narrative-explicit")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "stakeholder-classifications",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/stakeholder-classifications/"+opts.ID, jsonBody)
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

	row := buildStakeholderClassificationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated stakeholder classification %s (%s)\n", row.ID, row.Title)
	return nil
}

func parseDoStakeholderClassificationsUpdateOptions(cmd *cobra.Command, args []string) (doStakeholderClassificationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	title, _ := cmd.Flags().GetString("title")
	leverageFactor, _ := cmd.Flags().GetInt("leverage-factor")
	objectivesNarrativeExplicit, _ := cmd.Flags().GetString("objectives-narrative-explicit")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doStakeholderClassificationsUpdateOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		ID:                          args[0],
		Title:                       title,
		LeverageFactor:              leverageFactor,
		ObjectivesNarrativeExplicit: objectivesNarrativeExplicit,
	}, nil
}
