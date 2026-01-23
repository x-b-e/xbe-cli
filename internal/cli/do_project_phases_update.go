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

type doProjectPhasesUpdateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	ID               string
	Name             string
	Description      string
	SequencePosition int
}

func newDoProjectPhasesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project phase",
		Long: `Update a project phase.

Optional:
  --name             Phase name
  --description      Phase description
  --sequence-position  Position in sequence`,
		Example: `  # Update phase name
  xbe do project-phases update 123 --name "Phase 1 - Updated"

  # Update description
  xbe do project-phases update 123 --description "New description"

  # Update sequence position
  xbe do project-phases update 123 --sequence-position 3`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectPhasesUpdate,
	}
	initDoProjectPhasesUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectPhasesCmd.AddCommand(newDoProjectPhasesUpdateCmd())
}

func initDoProjectPhasesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Phase name")
	cmd.Flags().String("description", "", "Phase description")
	cmd.Flags().Int("sequence-position", 0, "Position in sequence")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectPhasesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectPhasesUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("sequence-position") {
		attributes["sequence-position"] = opts.SequencePosition
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "project-phases",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/project-phases/"+opts.ID, jsonBody)
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

	if opts.JSON {
		row := projectPhaseRow{
			ID:               resp.Data.ID,
			Name:             stringAttr(resp.Data.Attributes, "name"),
			Description:      stringAttr(resp.Data.Attributes, "description"),
			Sequence:         stringAttr(resp.Data.Attributes, "sequence"),
			SequencePosition: intAttr(resp.Data.Attributes, "sequence-position"),
		}
		if rel, ok := resp.Data.Relationships["project"]; ok && rel.Data != nil {
			row.ProjectID = rel.Data.ID
		}
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project phase %s\n", resp.Data.ID)
	return nil
}

func parseDoProjectPhasesUpdateOptions(cmd *cobra.Command, args []string) (doProjectPhasesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	sequencePosition, _ := cmd.Flags().GetInt("sequence-position")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectPhasesUpdateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		ID:               args[0],
		Name:             name,
		Description:      description,
		SequencePosition: sequencePosition,
	}, nil
}
