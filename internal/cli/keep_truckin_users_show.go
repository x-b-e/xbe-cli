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

type keepTruckinUsersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type keepTruckinUserDetails struct {
	ID           string `json:"id"`
	DriverID     string `json:"driver_id,omitempty"`
	EmailAddress string `json:"email_address,omitempty"`
	MobileNumber string `json:"mobile_number,omitempty"`
	FirstName    string `json:"first_name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
	Role         string `json:"role,omitempty"`
	Active       bool   `json:"active"`
	CarrierName  string `json:"carrier_name,omitempty"`
	UserSetAt    string `json:"user_set_at,omitempty"`
	BrokerID     string `json:"broker_id,omitempty"`
	TruckerID    string `json:"trucker_id,omitempty"`
	UserID       string `json:"user_id,omitempty"`
}

func newKeepTruckinUsersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show KeepTruckin user details",
		Long: `Show the full details of a KeepTruckin user.

Output Fields:
  ID
  Driver ID
  Email Address
  Mobile Number
  First Name
  Last Name
  Role
  Active
  Carrier Name
  User Set At
  Broker ID
  Trucker ID
  User ID

Arguments:
  <id>    The KeepTruckin user ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a KeepTruckin user
  xbe view keep-truckin-users show 123

  # Get JSON output
  xbe view keep-truckin-users show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runKeepTruckinUsersShow,
	}
	initKeepTruckinUsersShowFlags(cmd)
	return cmd
}

func init() {
	keepTruckinUsersCmd.AddCommand(newKeepTruckinUsersShowCmd())
}

func initKeepTruckinUsersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runKeepTruckinUsersShow(cmd *cobra.Command, args []string) error {
	opts, err := parseKeepTruckinUsersShowOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("keep-truckin user id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[keep-truckin-users]", "driver-id,email-address,mobile-number,first-name,last-name,role,active,carrier-name,user-set-at,broker,trucker,user")

	body, _, err := client.Get(cmd.Context(), "/v1/keep-truckin-users/"+id, query)
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

	details := buildKeepTruckinUserDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderKeepTruckinUserDetails(cmd, details)
}

func parseKeepTruckinUsersShowOptions(cmd *cobra.Command) (keepTruckinUsersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return keepTruckinUsersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildKeepTruckinUserDetails(resp jsonAPISingleResponse) keepTruckinUserDetails {
	attrs := resp.Data.Attributes
	return keepTruckinUserDetails{
		ID:           resp.Data.ID,
		DriverID:     stringAttr(attrs, "driver-id"),
		EmailAddress: stringAttr(attrs, "email-address"),
		MobileNumber: stringAttr(attrs, "mobile-number"),
		FirstName:    stringAttr(attrs, "first-name"),
		LastName:     stringAttr(attrs, "last-name"),
		Role:         stringAttr(attrs, "role"),
		Active:       boolAttr(attrs, "active"),
		CarrierName:  stringAttr(attrs, "carrier-name"),
		UserSetAt:    formatDateTime(stringAttr(attrs, "user-set-at")),
		BrokerID:     relationshipIDFromMap(resp.Data.Relationships, "broker"),
		TruckerID:    relationshipIDFromMap(resp.Data.Relationships, "trucker"),
		UserID:       relationshipIDFromMap(resp.Data.Relationships, "user"),
	}
}

func renderKeepTruckinUserDetails(cmd *cobra.Command, details keepTruckinUserDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.DriverID != "" {
		fmt.Fprintf(out, "Driver ID: %s\n", details.DriverID)
	}
	if details.EmailAddress != "" {
		fmt.Fprintf(out, "Email Address: %s\n", details.EmailAddress)
	}
	if details.MobileNumber != "" {
		fmt.Fprintf(out, "Mobile Number: %s\n", details.MobileNumber)
	}
	if details.FirstName != "" {
		fmt.Fprintf(out, "First Name: %s\n", details.FirstName)
	}
	if details.LastName != "" {
		fmt.Fprintf(out, "Last Name: %s\n", details.LastName)
	}
	if details.Role != "" {
		fmt.Fprintf(out, "Role: %s\n", details.Role)
	}
	fmt.Fprintf(out, "Active: %t\n", details.Active)
	if details.CarrierName != "" {
		fmt.Fprintf(out, "Carrier Name: %s\n", details.CarrierName)
	}
	if details.UserSetAt != "" {
		fmt.Fprintf(out, "User Set At: %s\n", details.UserSetAt)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.TruckerID != "" {
		fmt.Fprintf(out, "Trucker ID: %s\n", details.TruckerID)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User ID: %s\n", details.UserID)
	}

	return nil
}
