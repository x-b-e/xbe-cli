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

type driverDayConstraintsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type driverDayConstraintDetails struct {
	ID           string `json:"id"`
	DriverDayID  string `json:"driver_day_id,omitempty"`
	ConstraintID string `json:"constraint_id,omitempty"`
}

func newDriverDayConstraintsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show driver day constraint details",
		Long: `Show the full details of a driver day constraint.

Output Fields:
  ID
  Driver Day ID
  Constraint ID

Arguments:
  <id>    The driver day constraint ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a driver day constraint
  xbe view driver-day-constraints show 123

  # Output as JSON
  xbe view driver-day-constraints show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDriverDayConstraintsShow,
	}
	initDriverDayConstraintsShowFlags(cmd)
	return cmd
}

func init() {
	driverDayConstraintsCmd.AddCommand(newDriverDayConstraintsShowCmd())
}

func initDriverDayConstraintsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDriverDayConstraintsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseDriverDayConstraintsShowOptions(cmd)
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
		return fmt.Errorf("driver day constraint id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "driver-day,constraint")

	body, _, err := client.Get(cmd.Context(), "/v1/driver-day-constraints/"+id, query)
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

	details := buildDriverDayConstraintDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderDriverDayConstraintDetails(cmd, details)
}

func parseDriverDayConstraintsShowOptions(cmd *cobra.Command) (driverDayConstraintsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return driverDayConstraintsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildDriverDayConstraintDetails(resp jsonAPISingleResponse) driverDayConstraintDetails {
	resource := resp.Data
	details := driverDayConstraintDetails{
		ID: resource.ID,
	}
	if rel, ok := resource.Relationships["driver-day"]; ok && rel.Data != nil {
		details.DriverDayID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["constraint"]; ok && rel.Data != nil {
		details.ConstraintID = rel.Data.ID
	}
	return details
}

func renderDriverDayConstraintDetails(cmd *cobra.Command, details driverDayConstraintDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.DriverDayID != "" {
		fmt.Fprintf(out, "Driver Day ID: %s\n", details.DriverDayID)
	}
	if details.ConstraintID != "" {
		fmt.Fprintf(out, "Constraint ID: %s\n", details.ConstraintID)
	}

	return nil
}
