#!/bin/bash
#
# XBE CLI Integration Test Configuration
#
# This file configures the test environment. Tests MUST run against staging.
#

# Default to staging server
export XBE_BASE_URL="${XBE_BASE_URL:-https://server-staging.x-b-e.com}"

# Token must be provided via environment or stored auth
export XBE_TOKEN="${XBE_TOKEN:-}"

# Test configuration
export XBE_TEST_TIMEOUT="${XBE_TEST_TIMEOUT:-30}"  # seconds per command
export XBE_TEST_VERBOSE="${XBE_TEST_VERBOSE:-0}"   # set to 1 for verbose output

# ============================================================================
# Safety Checks
# ============================================================================

check_config() {
    # Ensure we're running against staging
    if [[ "$XBE_BASE_URL" != *"staging"* ]]; then
        echo ""
        echo "============================================"
        echo "ERROR: Integration tests MUST run against staging!"
        echo ""
        echo "Current base URL: $XBE_BASE_URL"
        echo ""
        echo "To run tests safely, either:"
        echo "  1. Use the default staging URL (recommended)"
        echo "  2. Set XBE_BASE_URL to contain 'staging'"
        echo ""
        echo "Example:"
        echo "  XBE_BASE_URL=https://server-staging.x-b-e.com ./tests/run_tests.sh"
        echo "============================================"
        exit 1
    fi

    # Check for token
    if [[ -z "$XBE_TOKEN" ]]; then
        # Try to get token from stored auth
        echo "No XBE_TOKEN provided, checking stored authentication..."

        # Check if xbe binary exists to test stored auth
        if [[ -x "${PROJECT_ROOT:-$(pwd)}/xbe" ]]; then
            local auth_status
            auth_status=$("${PROJECT_ROOT:-$(pwd)}/xbe" auth status --base-url "$XBE_BASE_URL" 2>&1) || true

            if [[ "$auth_status" == *"Token: set"* ]]; then
                echo "Using stored authentication (keychain)."
                # Token will be resolved by the CLI - set marker so test_helpers knows
                export XBE_USE_STORED_AUTH="1"
            else
                echo ""
                echo "============================================"
                echo "ERROR: No authentication found!"
                echo ""
                echo "Either:"
                echo "  1. Set XBE_TOKEN environment variable"
                echo "  2. Run: ./xbe auth login --base-url $XBE_BASE_URL"
                echo ""
                echo "To get a token for staging, use the XBE staging web UI"
                echo "or contact your administrator."
                echo "============================================"
                exit 1
            fi
        else
            echo ""
            echo "============================================"
            echo "ERROR: XBE_TOKEN is required!"
            echo ""
            echo "Set the XBE_TOKEN environment variable:"
            echo "  export XBE_TOKEN=xbe_at_your_token_here"
            echo "  ./tests/run_tests.sh"
            echo ""
            echo "Or pass it inline:"
            echo "  XBE_TOKEN=xbe_at_xxx ./tests/run_tests.sh"
            echo "============================================"
            exit 1
        fi
    fi

    # Print configuration
    echo ""
    echo "Test Configuration:"
    echo "  Base URL: $XBE_BASE_URL"
    if [[ -n "$XBE_TOKEN" ]]; then
        echo "  Token:    ${XBE_TOKEN:0:15}..."
    else
        echo "  Token:    (using stored auth)"
    fi
    echo ""
}

