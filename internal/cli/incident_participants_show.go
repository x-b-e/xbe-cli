package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type incidentParticipantsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type incidentParticipantDetails struct {
	ID                    string `json:"id"`
	Name                  string `json:"name,omitempty"`
	EmailAddress          string `json:"email_address,omitempty"`
	MobileNumber          string `json:"mobile_number,omitempty"`
	MobileNumberFormatted string `json:"mobile_number_formatted,omitempty"`
	Involvement           string `json:"involvement,omitempty"`
	IncidentID            string `json:"incident_id,omitempty"`
	UserID                string `json:"user_id,omitempty"`
}

func newIncidentParticipantsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show incident participant details",
		Long: `Show the full details of an incident participant.

Output Fields:
  ID
  Name
  Email Address
  Mobile Number
  Mobile Number Formatted
  Involvement
  Incident ID
  User ID

Arguments:
  <id>    The incident participant ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an incident participant
  xbe view incident-participants show 123

  # JSON output
  xbe view incident-participants show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runIncidentParticipantsShow,
	}
	initIncidentParticipantsShowFlags(cmd)
	return cmd
}

func init() {
	incidentParticipantsCmd.AddCommand(newIncidentParticipantsShowCmd())
}

func initIncidentParticipantsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runIncidentParticipantsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseIncidentParticipantsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("incident participant id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[incident-participants]", "name,email-address,mobile-number,mobile-number-formatted,involvement,incident,user")

	body, _, err := client.Get(cmd.Context(), "/v1/incident-participants/"+id, query)
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

	return renderIncidentParticipantDetails(cmd, details)
}

func parseIncidentParticipantsShowOptions(cmd *cobra.Command) (incidentParticipantsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return incidentParticipantsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildIncidentParticipantDetails(resp jsonAPISingleResponse) incidentParticipantDetails {
	attrs := resp.Data.Attributes
	return incidentParticipantDetails{
		ID:                    resp.Data.ID,
		Name:                  stringAttr(attrs, "name"),
		EmailAddress:          stringAttr(attrs, "email-address"),
		MobileNumber:          stringAttr(attrs, "mobile-number"),
		MobileNumberFormatted: stringAttr(attrs, "mobile-number-formatted"),
		Involvement:           stringAttr(attrs, "involvement"),
		IncidentID:            relationshipIDFromMap(resp.Data.Relationships, "incident"),
		UserID:                relationshipIDFromMap(resp.Data.Relationships, "user"),
	}
}

func renderIncidentParticipantDetails(cmd *cobra.Command, details incidentParticipantDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}
	if details.EmailAddress != "" {
		fmt.Fprintf(out, "Email Address: %s\n", details.EmailAddress)
	}
	if details.MobileNumber != "" {
		fmt.Fprintf(out, "Mobile Number: %s\n", details.MobileNumber)
	}
	if details.MobileNumberFormatted != "" {
		fmt.Fprintf(out, "Mobile Number Formatted: %s\n", details.MobileNumberFormatted)
	}
	if details.Involvement != "" {
		fmt.Fprintf(out, "Involvement: %s\n", details.Involvement)
	}
	if details.IncidentID != "" {
		fmt.Fprintf(out, "Incident ID: %s\n", details.IncidentID)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User ID: %s\n", details.UserID)
	}

	return nil
}
