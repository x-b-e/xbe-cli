# XBE CLI Integration Tests

Automated integration tests for the XBE CLI that exercise all commands against the staging server.

## Prerequisites

1. **jq** - JSON processor (required for parsing JSON output)
   ```bash
   brew install jq
   ```

2. **XBE CLI binary** - Built from source
   ```bash
   make build
   ```

3. **Staging authentication** - Either:
   - Set `XBE_TOKEN` environment variable, or
   - Run `./xbe auth login --base-url https://server-staging.x-b-e.com`

## Quick Start

```bash
# Run all tests sequentially
./tests/run_tests.sh

# Run all tests in parallel (8 jobs by default)
./tests/run_tests.sh -p

# Run all tests in parallel with custom job count
./tests/run_tests.sh -p -j 4

# Run with explicit token
XBE_TOKEN=xbe_at_xxx ./tests/run_tests.sh

# Run specific resource tests
./tests/run_tests.sh users
./tests/run_tests.sh brokers customers

# Run specific tests in parallel
./tests/run_tests.sh -p users brokers customers

# Run a single test file directly
./tests/resources/users_test.sh
```

## Parallel Execution

Use `-p` or `--parallel` to run test suites concurrently:

```bash
./tests/run_tests.sh -p           # 8 parallel jobs (default)
./tests/run_tests.sh -p -j 4      # 4 parallel jobs
./tests/run_tests.sh --parallel --jobs 12   # 12 parallel jobs
```

Each test suite runs independently with its own output captured to a temp file. Results are summarized at the end, with failed test output shown for debugging.

## Safety

**IMPORTANT:** Tests MUST run against staging. The test suite will refuse to run against any URL that doesn't contain "staging".

Tests automatically clean up created resources after completion.

## Directory Structure

```
tests/
├── run_tests.sh              # Main test runner
├── lib/
│   ├── test_helpers.sh       # Shared test functions and assertions
│   └── config.sh             # Environment configuration
├── resources/
│   ├── users_test.sh         # Users CRUD tests
│   ├── brokers_test.sh       # Brokers CRUD tests
│   └── ...                   # One file per resource
└── README.md                 # This file
```

## Test Structure

Each resource test file follows this pattern:

1. **Prerequisites** - Create required dependent resources (e.g., broker for customers)
2. **CREATE tests** - Test creating resources with various options
3. **UPDATE tests** - Test updating each attribute individually
4. **LIST tests** - Test listing with filters and pagination
5. **DELETE tests** - Test deletion with --confirm flag
6. **Error cases** - Test expected failures

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `XBE_BASE_URL` | API base URL | `https://server-staging.x-b-e.com` |
| `XBE_TOKEN` | API authentication token | (uses stored auth) |
| `XBE_TEST_VERBOSE` | Enable verbose output | `0` |

## Test Helpers

The test framework provides these assertion functions:

```bash
# Status assertions
assert_success          # Command exited with 0
assert_failure          # Command exited with non-zero

# Output assertions
assert_output_contains "text"     # Output contains substring
assert_output_not_contains "text" # Output doesn't contain substring

# JSON assertions
assert_json_has ".id"             # JSON has key at path
assert_json_equals ".name" "Bob"  # JSON value equals expected
assert_json_is_array              # JSON is an array
assert_json_array_not_empty       # JSON array has items
assert_json_bool ".active" "true" # JSON boolean value

# Utility functions
json_get ".id"          # Extract value from JSON output
unique_name "Prefix"    # Generate unique test name
unique_email            # Generate unique test email
unique_suffix           # Generate unique suffix
register_cleanup "resource" "$ID"  # Register resource for cleanup
```

## Running Commands

```bash
# Basic command execution
run ./xbe view users list --json

# XBE command with staging URL and token
xbe_run view users list --limit 5

# XBE command with --json flag
xbe_json view users list --limit 5
```

## Adding New Tests

1. Create a new file: `tests/resources/<resource>_test.sh`
2. Start with:
   ```bash
   #!/bin/bash
   source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

   describe "Resource: <resource>"
   ```
3. Add tests following the existing patterns
4. End with:
   ```bash
   run_tests
   ```
5. Make executable: `chmod +x tests/resources/<resource>_test.sh`

## Example Test

```bash
#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: widgets"

# Create
test_name "Create widget with name"
xbe_json do widgets create --name "Test Widget"
assert_success
WIDGET_ID=$(json_get ".id")
register_cleanup "widgets" "$WIDGET_ID"

# Update
test_name "Update widget name"
xbe_json do widgets update "$WIDGET_ID" --name "Updated Widget"
assert_success

# List
test_name "List widgets"
xbe_json view widgets list --limit 5
assert_json_is_array

# Delete
test_name "Delete widget"
xbe_run do widgets delete "$WIDGET_ID" --confirm
assert_success

run_tests
```

## Troubleshooting

### "Tests must run against staging server"
Ensure `XBE_BASE_URL` contains "staging" or use the default URL.

### "Authentication required"
Either set `XBE_TOKEN` or run `./xbe auth login --base-url https://server-staging.x-b-e.com`.

### "jq is required but not installed"
Install jq: `brew install jq`

### Tests fail to create resources
Check that your token has write permissions in staging.

### Cleanup fails
Some resources may have dependencies preventing deletion. Check staging manually.

## Resources Tested

### CRUD Resources (with create/update/delete)
- action-items
- brokers
- business-units
- certification-requirements
- certification-types
- certifications
- cost-codes
- cost-index-entries
- cost-indexes
- craft-classes
- crafts
- culture-values
- custom-work-order-statuses
- customers
- developer-reference-types
- developer-trucker-certification-classifications
- developers
- equipment-classifications
- external-identification-types
- glossary-terms
- job-production-plan-cancellation-reason-types
- job-sites
- labor-classifications
- material-sites
- material-suppliers
- memberships
- project-cost-classifications
- project-offices
- project-resource-classifications
- project-revenue-classifications
- project-transport-event-types
- projects
- quality-control-classifications
- shift-feedback-reasons
- stakeholder-classifications
- tag-categories
- tags
- time-sheet-line-item-classifications
- tractor-credentials
- tractor-trailer-credential-classifications
- tractors
- trailer-credentials
- trailers
- truck-scopes
- truckers
- user-credential-classifications
- user-credentials
- users

### View-Only Resources (list only)
- features
- incident-tags
- job-production-plan-inspectable-summaries
- job-production-plans
- languages
- material-transactions
- material-types
- newsletters
- posts
- post-router-jobs
- press-releases
- profit-improvement-categories
- project-categories
- project-divisions
- reaction-classifications
- release-notes
- service-types
- trailer-classifications
- transport-orders
- unit-of-measures
