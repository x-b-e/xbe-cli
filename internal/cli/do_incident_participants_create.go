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

type doIncidentParticipantsCreateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	Incident     string
	Name         string
	EmailAddress string
	MobileNumber string
	Involvement  string
}

func newDoIncidentParticipantsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an incident participant",
		Long: `Create an incident participant.

Required flags:
  --incident       Incident ID
  --name           Participant name
  --email-address  Email address (required if mobile number is blank)
  --mobile-number  Mobile number (required if email address is blank)

Optional flags:
  --involvement    Involvement summary

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create an incident participant with email
  xbe do incident-participants create \
    --incident 123 \
    --name "Jamie Doe" \
    --email-address "jamie@example.com" \
    --involvement witness

  # Create an incident participant with mobile number
  xbe do incident-participants create \
    --incident 123 \
    --name "Alex Doe" \
    --mobile-number "+18153471234"

  # Get JSON output
  xbe do incident-participants create \
    --incident 123 \
    --name "Jamie Doe" \
    --email-address "jamie@example.com" \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoIncidentParticipantsCreate,
	}
	initDoIncidentParticipantsCreateFlags(cmd)
	return cmd
}

func init() {
	doIncidentParticipantsCmd.AddCommand(newDoIncidentParticipantsCreateCmd())
}

func initDoIncidentParticipantsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("incident", "", "Incident ID (required)")
	cmd.Flags().String("name", "", "Participant name (required)")
	cmd.Flags().String("email-address", "", "Email address")
	cmd.Flags().String("mobile-number", "", "Mobile number")
	cmd.Flags().String("involvement", "", "Involvement summary")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoIncidentParticipantsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoIncidentParticipantsCreateOptions(cmd)
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

	incidentID := strings.TrimSpace(opts.Incident)
	if incidentID == "" {
		err := fmt.Errorf("--incident is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	name := strings.TrimSpace(opts.Name)
	if name == "" {
		err := fmt.Errorf("--name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	emailAddress := strings.TrimSpace(opts.EmailAddress)
	mobileNumber := strings.TrimSpace(opts.MobileNumber)
	if emailAddress == "" && mobileNumber == "" {
		err := fmt.Errorf("--email-address or --mobile-number is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"name": name,
	}
	if emailAddress != "" {
		attributes["email-address"] = emailAddress
	}
	if mobileNumber != "" {
		attributes["mobile-number"] = mobileNumber
	}
	if strings.TrimSpace(opts.Involvement) != "" {
		attributes["involvement"] = opts.Involvement
	}

	relationships := map[string]any{
		"incident": map[string]any{
			"data": map[string]any{
				"type": "incidents",
				"id":   incidentID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "incident-participants",
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

	body, _, err := client.Post(cmd.Context(), "/v1/incident-participants", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created incident participant %s\n", details.ID)
	return nil
}

func parseDoIncidentParticipantsCreateOptions(cmd *cobra.Command) (doIncidentParticipantsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	incident, _ := cmd.Flags().GetString("incident")
	name, _ := cmd.Flags().GetString("name")
	emailAddress, _ := cmd.Flags().GetString("email-address")
	mobileNumber, _ := cmd.Flags().GetString("mobile-number")
	involvement, _ := cmd.Flags().GetString("involvement")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doIncidentParticipantsCreateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		Incident:     incident,
		Name:         name,
		EmailAddress: emailAddress,
		MobileNumber: mobileNumber,
		Involvement:  involvement,
	}, nil
}
