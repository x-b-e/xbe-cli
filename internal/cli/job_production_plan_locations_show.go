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

type jobProductionPlanLocationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

func newJobProductionPlanLocationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan location details",
		Long: `Show the full details of a job production plan location.

Output Fields:
  ID
  Name
  Site Kind
  Start Site Candidate
  Job Production Plan ID
  Segment ID
  Address
  Address Formatted
  Address Time Zone ID
  Address City
  Address State Code
  Address Latitude
  Address Longitude
  Address Place ID
  Address Plus Code

Arguments:
  <id>    The job production plan location ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a job production plan location
  xbe view job-production-plan-locations show 123

  # Output as JSON
  xbe view job-production-plan-locations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanLocationsShow,
	}
	initJobProductionPlanLocationsShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanLocationsCmd.AddCommand(newJobProductionPlanLocationsShowCmd())
}

func initJobProductionPlanLocationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanLocationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobProductionPlanLocationsShowOptions(cmd)
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
		return fmt.Errorf("job production plan location id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[job-production-plan-locations]", "name,site-kind,is-start-site-candidate,address,is-address-formatted-address,address-formatted,address-time-zone-id,address-city,address-state-code,address-latitude,address-longitude,address-place-id,address-plus-code,job-production-plan,segment")

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-locations/"+id, query)
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

	details := buildJobProductionPlanLocationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanLocationDetails(cmd, details)
}

func parseJobProductionPlanLocationsShowOptions(cmd *cobra.Command) (jobProductionPlanLocationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanLocationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func renderJobProductionPlanLocationDetails(cmd *cobra.Command, details jobProductionPlanLocationRow) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}
	if details.SiteKind != "" {
		fmt.Fprintf(out, "Site Kind: %s\n", details.SiteKind)
	}
	fmt.Fprintf(out, "Start Site Candidate: %t\n", details.IsStartSiteCandidate)
	if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan ID: %s\n", details.JobProductionPlanID)
	}
	if details.SegmentID != "" {
		fmt.Fprintf(out, "Segment ID: %s\n", details.SegmentID)
	}
	if details.Address != "" {
		fmt.Fprintf(out, "Address: %s\n", details.Address)
	}
	if details.AddressFormatted != "" {
		fmt.Fprintf(out, "Address Formatted: %s\n", details.AddressFormatted)
	}
	if details.AddressTimeZoneID != "" {
		fmt.Fprintf(out, "Address Time Zone ID: %s\n", details.AddressTimeZoneID)
	}
	if details.AddressCity != "" {
		fmt.Fprintf(out, "Address City: %s\n", details.AddressCity)
	}
	if details.AddressStateCode != "" {
		fmt.Fprintf(out, "Address State Code: %s\n", details.AddressStateCode)
	}
	if details.AddressLatitude != "" {
		fmt.Fprintf(out, "Address Latitude: %s\n", details.AddressLatitude)
	}
	if details.AddressLongitude != "" {
		fmt.Fprintf(out, "Address Longitude: %s\n", details.AddressLongitude)
	}
	if details.AddressPlaceID != "" {
		fmt.Fprintf(out, "Address Place ID: %s\n", details.AddressPlaceID)
	}
	if details.AddressPlusCode != "" {
		fmt.Fprintf(out, "Address Plus Code: %s\n", details.AddressPlusCode)
	}

	return nil
}
