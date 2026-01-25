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

type doOpenDoorIssuesUpdateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	ID           string
	Description  string
	Status       string
	Organization string
	ReportedBy   string
}

func newDoOpenDoorIssuesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an open door issue",
		Long: `Update an open door issue.

Optional flags:
  --description   Issue description
  --status        Issue status (editing, reported, resolved)
  --organization  Organization in Type|ID format (e.g. Broker|123)
  --reported-by   Reporting user ID`,
		Example: `  # Update the status
  xbe do open-door-issues update 123 --status resolved

  # Update the description
  xbe do open-door-issues update 123 --description "Updated details"

  # Change organization and reporter
  xbe do open-door-issues update 123 --organization "Broker|456" --reported-by 789`,
		Args: cobra.ExactArgs(1),
		RunE: runDoOpenDoorIssuesUpdate,
	}
	initDoOpenDoorIssuesUpdateFlags(cmd)
	return cmd
}

func init() {
	doOpenDoorIssuesCmd.AddCommand(newDoOpenDoorIssuesUpdateCmd())
}

func initDoOpenDoorIssuesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("description", "", "Issue description")
	cmd.Flags().String("status", "", "Issue status (editing, reported, resolved)")
	cmd.Flags().String("organization", "", "Organization in Type|ID format (e.g. Broker|123)")
	cmd.Flags().String("reported-by", "", "Reporting user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoOpenDoorIssuesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoOpenDoorIssuesUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("organization") {
		if opts.Organization == "" {
			err := fmt.Errorf("--organization cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		orgType, orgID, err := parseOrganization(opts.Organization)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["organization"] = map[string]any{
			"data": map[string]any{
				"type": orgType,
				"id":   orgID,
			},
		}
	}
	if cmd.Flags().Changed("reported-by") {
		if opts.ReportedBy == "" {
			err := fmt.Errorf("--reported-by cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["reported-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.ReportedBy,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "open-door-issues",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/open-door-issues/"+opts.ID, jsonBody)
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

	row := buildOpenDoorIssueRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated open door issue %s\n", row.ID)
	return nil
}

func parseDoOpenDoorIssuesUpdateOptions(cmd *cobra.Command, args []string) (doOpenDoorIssuesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	description, _ := cmd.Flags().GetString("description")
	status, _ := cmd.Flags().GetString("status")
	organization, _ := cmd.Flags().GetString("organization")
	reportedBy, _ := cmd.Flags().GetString("reported-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doOpenDoorIssuesUpdateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		ID:           args[0],
		Description:  description,
		Status:       status,
		Organization: organization,
		ReportedBy:   reportedBy,
	}, nil
}
