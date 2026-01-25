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

type rawTransportDriversShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type rawTransportDriverDetails struct {
	ID                  string   `json:"id"`
	ExternalDriverID    string   `json:"external_driver_id,omitempty"`
	Importer            string   `json:"importer,omitempty"`
	ImportStatus        string   `json:"import_status,omitempty"`
	ImportErrors        []string `json:"import_errors,omitempty"`
	BrokerID            string   `json:"broker_id,omitempty"`
	UserID              string   `json:"user_id,omitempty"`
	TruckerMembershipID string   `json:"trucker_membership_id,omitempty"`
	Tables              any      `json:"tables,omitempty"`
}

func newRawTransportDriversShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show raw transport driver details",
		Long: `Show the full details of a raw transport driver.

Output Fields:
  ID
  External Driver ID
  Importer
  Import Status
  Import Errors
  Broker
  User
  Trucker Membership
  Tables

Arguments:
  <id>    The raw transport driver ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show raw transport driver details
  xbe view raw-transport-drivers show 123

  # Output as JSON
  xbe view raw-transport-drivers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runRawTransportDriversShow,
	}
	initRawTransportDriversShowFlags(cmd)
	return cmd
}

func init() {
	rawTransportDriversCmd.AddCommand(newRawTransportDriversShowCmd())
}

func initRawTransportDriversShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRawTransportDriversShow(cmd *cobra.Command, args []string) error {
	opts, err := parseRawTransportDriversShowOptions(cmd)
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
		return fmt.Errorf("raw transport driver id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[raw-transport-drivers]", "external-driver-id,importer,import-status,import-errors,tables")

	body, _, err := client.Get(cmd.Context(), "/v1/raw-transport-drivers/"+id, query)
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

	details := buildRawTransportDriverDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderRawTransportDriverDetails(cmd, details)
}

func parseRawTransportDriversShowOptions(cmd *cobra.Command) (rawTransportDriversShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return rawTransportDriversShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildRawTransportDriverDetails(resp jsonAPISingleResponse) rawTransportDriverDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := rawTransportDriverDetails{
		ID:                  resource.ID,
		ExternalDriverID:    stringAttr(attrs, "external-driver-id"),
		Importer:            stringAttr(attrs, "importer"),
		ImportStatus:        stringAttr(attrs, "import-status"),
		ImportErrors:        stringSliceAttr(attrs, "import-errors"),
		BrokerID:            relationshipIDFromMap(resource.Relationships, "broker"),
		UserID:              relationshipIDFromMap(resource.Relationships, "user"),
		TruckerMembershipID: relationshipIDFromMap(resource.Relationships, "trucker-membership"),
		Tables:              attrs["tables"],
	}

	return details
}

func renderRawTransportDriverDetails(cmd *cobra.Command, details rawTransportDriverDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ExternalDriverID != "" {
		fmt.Fprintf(out, "External Driver ID: %s\n", details.ExternalDriverID)
	}
	if details.Importer != "" {
		fmt.Fprintf(out, "Importer: %s\n", details.Importer)
	}
	if details.ImportStatus != "" {
		fmt.Fprintf(out, "Import Status: %s\n", details.ImportStatus)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User ID: %s\n", details.UserID)
	}
	if details.TruckerMembershipID != "" {
		fmt.Fprintf(out, "Trucker Membership ID: %s\n", details.TruckerMembershipID)
	}
	if len(details.ImportErrors) > 0 {
		fmt.Fprintf(out, "Import Errors: %s\n", strings.Join(details.ImportErrors, ", "))
	}
	if details.Tables != nil {
		fmt.Fprintln(out, "\nTables:")
		fmt.Fprintln(out, formatJSONBlock(details.Tables, "  "))
	}

	return nil
}
