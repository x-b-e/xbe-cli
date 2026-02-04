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

type promptPrescriptionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type promptPrescriptionDetails struct {
	ID               string   `json:"id"`
	Name             string   `json:"name,omitempty"`
	EmailAddress     string   `json:"email_address,omitempty"`
	OrganizationName string   `json:"organization_name,omitempty"`
	LocationName     string   `json:"location_name,omitempty"`
	Role             string   `json:"role,omitempty"`
	Symptoms         string   `json:"symptoms,omitempty"`
	Prompts          []string `json:"prompts,omitempty"`
	CreatedAt        string   `json:"created_at,omitempty"`
	UpdatedAt        string   `json:"updated_at,omitempty"`
}

func newPromptPrescriptionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show prompt prescription details",
		Long: `Show the full details of a prompt prescription.

Output Fields:
  ID
  Name
  Email Address
  Organization Name
  Location Name
  Role
  Symptoms
  Prompts
  Created At
  Updated At

Arguments:
  <id>  The prompt prescription ID (required).`,
		Example: `  # Show a prompt prescription
  xbe view prompt-prescriptions show 123

  # Output as JSON
  xbe view prompt-prescriptions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runPromptPrescriptionsShow,
	}
	initPromptPrescriptionsShowFlags(cmd)
	return cmd
}

func init() {
	promptPrescriptionsCmd.AddCommand(newPromptPrescriptionsShowCmd())
}

func initPromptPrescriptionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPromptPrescriptionsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parsePromptPrescriptionsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("prompt prescription id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[prompt-prescriptions]", "name,email-address,organization-name,location-name,role,symptoms,prompts,created-at,updated-at")

	body, _, err := client.Get(cmd.Context(), "/v1/prompt-prescriptions/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildPromptPrescriptionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderPromptPrescriptionDetails(cmd, details)
}

func parsePromptPrescriptionsShowOptions(cmd *cobra.Command) (promptPrescriptionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return promptPrescriptionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildPromptPrescriptionDetails(resp jsonAPISingleResponse) promptPrescriptionDetails {
	attrs := resp.Data.Attributes

	return promptPrescriptionDetails{
		ID:               resp.Data.ID,
		Name:             strings.TrimSpace(stringAttr(attrs, "name")),
		EmailAddress:     strings.TrimSpace(stringAttr(attrs, "email-address")),
		OrganizationName: strings.TrimSpace(stringAttr(attrs, "organization-name")),
		LocationName:     strings.TrimSpace(stringAttr(attrs, "location-name")),
		Role:             strings.TrimSpace(stringAttr(attrs, "role")),
		Symptoms:         strings.TrimSpace(stringAttr(attrs, "symptoms")),
		Prompts:          stringSliceAttr(attrs, "prompts"),
		CreatedAt:        stringAttr(attrs, "created-at"),
		UpdatedAt:        stringAttr(attrs, "updated-at"),
	}
}

func renderPromptPrescriptionDetails(cmd *cobra.Command, details promptPrescriptionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}
	if details.EmailAddress != "" {
		fmt.Fprintf(out, "Email Address: %s\n", details.EmailAddress)
	}
	if details.OrganizationName != "" {
		fmt.Fprintf(out, "Organization Name: %s\n", details.OrganizationName)
	}
	if details.LocationName != "" {
		fmt.Fprintf(out, "Location Name: %s\n", details.LocationName)
	}
	if details.Role != "" {
		fmt.Fprintf(out, "Role: %s\n", details.Role)
	}
	if details.Symptoms != "" {
		fmt.Fprintf(out, "Symptoms: %s\n", details.Symptoms)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	if len(details.Prompts) > 0 {
		fmt.Fprintln(out, "Prompts:")
		for i, prompt := range details.Prompts {
			line := strings.TrimSpace(prompt)
			if line == "" {
				continue
			}
			fmt.Fprintf(out, "  %d. %s\n", i+1, line)
		}
	}

	return nil
}
