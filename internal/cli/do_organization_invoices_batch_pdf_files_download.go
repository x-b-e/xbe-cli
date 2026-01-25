package cli

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
	"github.com/xbe-inc/xbe-cli/internal/version"
)

type doOrganizationInvoicesBatchPdfFilesDownloadOptions struct {
	BaseURL   string
	Token     string
	ID        string
	Output    string
	Overwrite bool
}

func newDoOrganizationInvoicesBatchPdfFilesDownloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download <id>",
		Short: "Download an organization invoices batch PDF file",
		Long: `Download an organization invoices batch PDF file.

The PDF is saved to disk. If --output is not provided, the command uses the
filename from the Content-Disposition header or defaults to
organization-invoices-batch-pdf-file-<id>.pdf.

Arguments:
  <id>    Organization invoices batch PDF file ID (required)

Global flags (see xbe --help): --base-url, --token`,
		Example: `  # Download a PDF file to the default filename
  xbe do organization-invoices-batch-pdf-files download 123

  # Download to a specific path
  xbe do organization-invoices-batch-pdf-files download 123 --output ./invoice.pdf

  # Stream to stdout
  xbe do organization-invoices-batch-pdf-files download 123 --output -`,
		Args: cobra.ExactArgs(1),
		RunE: runDoOrganizationInvoicesBatchPdfFilesDownload,
	}
	initDoOrganizationInvoicesBatchPdfFilesDownloadFlags(cmd)
	return cmd
}

func init() {
	doOrganizationInvoicesBatchPdfFilesCmd.AddCommand(newDoOrganizationInvoicesBatchPdfFilesDownloadCmd())
}

func initDoOrganizationInvoicesBatchPdfFilesDownloadFlags(cmd *cobra.Command) {
	cmd.Flags().String("output", "", "Output file path (default: use server filename)")
	cmd.Flags().Bool("overwrite", false, "Overwrite existing file")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoOrganizationInvoicesBatchPdfFilesDownload(cmd *cobra.Command, args []string) error {
	opts, err := parseDoOrganizationInvoicesBatchPdfFilesDownloadOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			err := errors.New("authentication required. Run 'xbe auth login' first")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	reqURL, err := buildOrganizationInvoicesBatchPdfFilesDownloadURL(client.BaseURL, opts.ID)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	req, err := http.NewRequestWithContext(cmd.Context(), http.MethodGet, reqURL, nil)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if client.Token != "" {
		req.Header.Set("Authorization", "Bearer "+client.Token)
	}
	req.Header.Set("Accept", "application/pdf")
	req.Header.Set("User-Agent", "xbe-cli/"+version.String())

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		err := fmt.Errorf("request failed: %s", resp.Status)
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	outputPath := resolveDownloadOutputPath(opts.Output, resp.Header.Get("Content-Disposition"), opts.ID)
	if outputPath == "-" {
		_, err = io.Copy(cmd.OutOrStdout(), resp.Body)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		return nil
	}

	if outputPath == "" {
		err := errors.New("output path could not be determined")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.Overwrite {
		if _, err := os.Stat(outputPath); err == nil {
			err := fmt.Errorf("output file already exists: %s", outputPath)
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		} else if !errors.Is(err, os.ErrNotExist) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	file, err := os.Create(outputPath)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Saved PDF to %s\n", outputPath)
	return nil
}

func parseDoOrganizationInvoicesBatchPdfFilesDownloadOptions(cmd *cobra.Command, args []string) (doOrganizationInvoicesBatchPdfFilesDownloadOptions, error) {
	output, _ := cmd.Flags().GetString("output")
	overwrite, _ := cmd.Flags().GetBool("overwrite")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doOrganizationInvoicesBatchPdfFilesDownloadOptions{
		BaseURL:   baseURL,
		Token:     token,
		ID:        args[0],
		Output:    output,
		Overwrite: overwrite,
	}, nil
}

func buildOrganizationInvoicesBatchPdfFilesDownloadURL(baseURL, id string) (string, error) {
	base, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return "", fmt.Errorf("invalid base url: %w", err)
	}

	path := "/v1/organization-invoices-batch-pdf-files/" + id + "/download"
	base.Path = strings.TrimRight(base.Path, "/") + path

	return base.String(), nil
}

func resolveDownloadOutputPath(explicitPath, contentDisposition, id string) string {
	if explicitPath != "" {
		return explicitPath
	}

	filename := filenameFromContentDisposition(contentDisposition)
	if filename != "" {
		return filename
	}

	return fmt.Sprintf("organization-invoices-batch-pdf-file-%s.pdf", id)
}

func filenameFromContentDisposition(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}

	_, params, err := mime.ParseMediaType(header)
	if err == nil {
		if filename := strings.TrimSpace(params["filename"]); filename != "" {
			return filepath.Base(filename)
		}
		if filename := strings.TrimSpace(params["filename*"]); filename != "" {
			decoded := decodeRFC5987Value(filename)
			if decoded != "" {
				return filepath.Base(decoded)
			}
		}
	}

	return ""
}

func decodeRFC5987Value(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	if idx := strings.Index(value, "''"); idx != -1 {
		value = value[idx+2:]
	}
	decoded, err := url.PathUnescape(value)
	if err != nil {
		return value
	}
	return decoded
}
