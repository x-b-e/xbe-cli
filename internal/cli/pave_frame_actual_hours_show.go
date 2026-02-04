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

type paveFrameActualHoursShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

func newPaveFrameActualHoursShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show pave frame actual hour details",
		Long: `Show detailed information for a single pave frame actual hour.

Output Fields:
  ID             Record identifier
  Date           Date of the hour record
  Hour           Hour of day (0-23)
  Window         Window (day/night)
  Latitude       Latitude
  Longitude      Longitude
  Temp Min (F)   Minimum temperature in Fahrenheit
  Precip 1hr (in)  Precipitation in the last hour (inches)

Arguments:
  <id>  Pave frame actual hour ID (required).`,
		Example: `  # Show a pave frame actual hour
  xbe view pave-frame-actual-hours show 123

  # JSON output
  xbe view pave-frame-actual-hours show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runPaveFrameActualHoursShow,
	}
	initPaveFrameActualHoursShowFlags(cmd)
	return cmd
}

func init() {
	paveFrameActualHoursCmd.AddCommand(newPaveFrameActualHoursShowCmd())
}

func initPaveFrameActualHoursShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPaveFrameActualHoursShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parsePaveFrameActualHoursShowOptions(cmd)
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
		return fmt.Errorf("pave frame actual hour id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[pave-frame-actual-hours]", "date,latitude,longitude,hour,window,temp-min-f,precip-1hr-in")

	body, _, err := client.Get(cmd.Context(), "/v1/pave-frame-actual-hours/"+id, query)
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

	row := buildPaveFrameActualHourRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	return renderPaveFrameActualHourDetails(cmd, row)
}

func parsePaveFrameActualHoursShowOptions(cmd *cobra.Command) (paveFrameActualHoursShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return paveFrameActualHoursShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return paveFrameActualHoursShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return paveFrameActualHoursShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return paveFrameActualHoursShowOptions{}, err
	}

	return paveFrameActualHoursShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func renderPaveFrameActualHourDetails(cmd *cobra.Command, row paveFrameActualHourRow) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", row.ID)
	if row.Date != "" {
		fmt.Fprintf(out, "Date: %s\n", row.Date)
	}
	if row.Hour != "" {
		fmt.Fprintf(out, "Hour: %s\n", row.Hour)
	}
	if row.Window != "" {
		fmt.Fprintf(out, "Window: %s\n", row.Window)
	}
	if row.Latitude != "" {
		fmt.Fprintf(out, "Latitude: %s\n", row.Latitude)
	}
	if row.Longitude != "" {
		fmt.Fprintf(out, "Longitude: %s\n", row.Longitude)
	}
	if row.TempMinF != "" {
		fmt.Fprintf(out, "Temp Min (F): %s\n", row.TempMinF)
	}
	if row.Precip1hrIn != "" {
		fmt.Fprintf(out, "Precip 1hr (in): %s\n", row.Precip1hrIn)
	}

	return nil
}
