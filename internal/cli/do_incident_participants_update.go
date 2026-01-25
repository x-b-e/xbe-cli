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

type doIncidentParticipantsUpdateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	ID           string
	Name         string
	EmailAddress string
	MobileNumber string
	Involvement  string
}

func newDoIncidentParticipantsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an incident participant",
		Long: `Update an existing incident participant.

Arguments:
  <id>    The incident participant ID (required)

Flags:
  --name           Participant name
  --email-address  Email address
  --mobile-number  Mobile number
  --involvement    Involvement summary

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update participant name
  xbe do incident-participants update 123 --name "Updated Name"

  # Update contact details
  xbe do incident-participants update 123 --email-address "new@example.com" --mobile-number "+18153471234"

  # Get JSON output
  xbe do incident-participants update 123 --name "Updated Name" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoIncidentParticipantsUpdate,
	}
	initDoIncidentParticipantsUpdateFlags(cmd)
	return cmd
}

func init() {
	doIncidentParticipantsCmd.AddCommand(newDoIncidentParticipantsUpdateCmd())
}

func initDoIncidentParticipantsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Participant name")
	cmd.Flags().String("email-address", "", "Email address")
	cmd.Flags().String("mobile-number", "", "Mobile number")
	cmd.Flags().String("involvement", "", "Involvement summary")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoIncidentParticipantsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoIncidentParticipantsUpdateOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
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
	if cmd.Flags().Changed("email-address") {
		attributes["email-address"] = opts.EmailAddress
	}
	if cmd.Flags().Changed("mobile-number") {
		attributes["mobile-number"] = opts.MobileNumber
	}
	if cmd.Flags().Changed("involvement") {
		attributes["involvement"] = opts.Involvement
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":       "incident-participants",
		"id":         opts.ID,
		"attributes": attributes,
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/incident-participants/"+opts.ID, jsonBody)
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

	details := buildIncidentParticipantDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated incident participant %s\n", details.ID)
	return nil
}

func parseDoIncidentParticipantsUpdateOptions(cmd *cobra.Command, args []string) (doIncidentParticipantsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	emailAddress, _ := cmd.Flags().GetString("email-address")
	mobileNumber, _ := cmd.Flags().GetString("mobile-number")
	involvement, _ := cmd.Flags().GetString("involvement")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doIncidentParticipantsUpdateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		ID:           args[0],
		Name:         name,
		EmailAddress: emailAddress,
		MobileNumber: mobileNumber,
		Involvement:  involvement,
	}, nil
}
