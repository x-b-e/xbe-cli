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

type doBrokerTruckerRatingsDeleteOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Confirm bool
}

func newDoBrokerTruckerRatingsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a broker trucker rating",
		Long: `Delete a broker trucker rating.

Provide the broker trucker rating ID as an argument. The --confirm flag is required
to prevent accidental deletions.`,
		Example: `  # Delete a broker trucker rating
  xbe do broker-trucker-ratings delete 123 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: runDoBrokerTruckerRatingsDelete,
	}
	initDoBrokerTruckerRatingsDeleteFlags(cmd)
	return cmd
}

func init() {
	doBrokerTruckerRatingsCmd.AddCommand(newDoBrokerTruckerRatingsDeleteCmd())
}

func initDoBrokerTruckerRatingsDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBrokerTruckerRatingsDelete(cmd *cobra.Command, args []string) error {
	opts, err := parseDoBrokerTruckerRatingsDeleteOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.Confirm {
		err := fmt.Errorf("--confirm flag is required to delete a broker trucker rating")
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

	if strings.TrimSpace(opts.ID) == "" {
		return fmt.Errorf("broker trucker rating id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[broker-trucker-ratings]", "rating,broker,trucker")

	getBody, _, err := client.Get(cmd.Context(), "/v1/broker-trucker-ratings/"+opts.ID, query)
	if err != nil {
		if len(getBody) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(getBody))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var getResp jsonAPISingleResponse
	if err := json.Unmarshal(getBody, &getResp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	row := brokerTruckerRatingRowFromSingle(getResp)

	deleteBody, _, err := client.Delete(cmd.Context(), "/v1/broker-trucker-ratings/"+opts.ID)
	if err != nil {
		if len(deleteBody) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(deleteBody))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.BrokerID != "" && row.TruckerID != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Deleted broker trucker rating %s (broker %s, trucker %s)\n", row.ID, row.BrokerID, row.TruckerID)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted broker trucker rating %s\n", opts.ID)
	return nil
}

func parseDoBrokerTruckerRatingsDeleteOptions(cmd *cobra.Command, args []string) (doBrokerTruckerRatingsDeleteOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	confirm, _ := cmd.Flags().GetBool("confirm")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBrokerTruckerRatingsDeleteOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Confirm: confirm,
	}, nil
}
