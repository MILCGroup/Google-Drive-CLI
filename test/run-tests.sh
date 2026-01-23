#!/bin/bash
# Test runner script for Google Drive CLI

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "================================="
echo "Google Drive CLI Test Runner"
echo "================================="
echo

# Function to print colored status
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ PASSED${NC}: $2"
    else
        echo -e "${RED}✗ FAILED${NC}: $2"
    fi
}

# Function to run test suite
run_test_suite() {
    local name=$1
    local path=$2
    local args=$3
    
    echo -e "${YELLOW}Running: $name${NC}"
    if go test $args -v $path 2>&1 | tee /tmp/test-output.log; then
        print_status 0 "$name"
        return 0
    else
        print_status 1 "$name"
        return 1
    fi
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    echo "Please install Go from https://golang.org/dl/"
    exit 1
fi

# Parse command line arguments
MODE="unit"
VERBOSE=""
COVERAGE=""
SHORT=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --integration)
            MODE="integration"
            shift
            ;;
        --all)
            MODE="all"
            shift
            ;;
        --coverage)
            COVERAGE="-coverprofile=coverage.out"
            shift
            ;;
        -v|--verbose)
            VERBOSE="-v"
            shift
            ;;
        --short)
            SHORT="-short"
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--integration|--all] [--coverage] [-v|--verbose] [--short]"
            exit 1
            ;;
    esac
done

TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

echo "Test Mode: $MODE"
echo "================================="
echo

# Run unit tests and property tests
if [ "$MODE" == "unit" ] || [ "$MODE" == "all" ]; then
    echo "Running Unit Tests and Property Tests"
    echo "-------------------------------------"
    
    # API tests
    ((TOTAL_TESTS++))
    if run_test_suite "API Client Tests" "./internal/api" "$VERBOSE $COVERAGE $SHORT"; then
        ((PASSED_TESTS++))
    else
        ((FAILED_TESTS++))
    fi
    
    # Auth tests
    ((TOTAL_TESTS++))
    if run_test_suite "Auth Tests" "./internal/auth" "$VERBOSE $SHORT"; then
        ((PASSED_TESTS++))
    else
        ((FAILED_TESTS++))
    fi
    
    # Utils tests
    ((TOTAL_TESTS++))
    if run_test_suite "Utils Tests" "./internal/utils" "$VERBOSE $SHORT"; then
        ((PASSED_TESTS++))
    else
        ((FAILED_TESTS++))
    fi
    
    # Files tests
    ((TOTAL_TESTS++))
    if run_test_suite "Files Tests" "./internal/files" "$VERBOSE $SHORT"; then
        ((PASSED_TESTS++))
    else
        ((FAILED_TESTS++))
    fi
    
    echo
fi

# Run integration tests
if [ "$MODE" == "integration" ] || [ "$MODE" == "all" ]; then
    echo "Running Integration Tests"
    echo "-------------------------"
    
    # Check for required environment variables
    if [ -z "$TEST_PROFILE" ]; then
        echo -e "${YELLOW}Warning: TEST_PROFILE not set${NC}"
        echo "Some integration tests may be skipped"
        echo "Set TEST_PROFILE to your authenticated profile name"
        echo
    fi
    
    # Integration tests
    ((TOTAL_TESTS++))
    if run_test_suite "Auth Integration Tests" "./test/integration" "-tags=integration $VERBOSE $SHORT -run TestIntegration_Auth"; then
        ((PASSED_TESTS++))
    else
        ((FAILED_TESTS++))
    fi
    
    ((TOTAL_TESTS++))
    if run_test_suite "API Integration Tests" "./test/integration" "-tags=integration $VERBOSE $SHORT -run TestIntegration_API"; then
        ((PASSED_TESTS++))
    else
        ((FAILED_TESTS++))
    fi
    
    ((TOTAL_TESTS++))
    if run_test_suite "File Operations Integration Tests" "./test/integration" "-tags=integration $VERBOSE $SHORT -run TestIntegration_FileOperations"; then
        ((PASSED_TESTS++))
    else
        ((FAILED_TESTS++))
    fi
    
    echo
fi

# Summary
echo "================================="
echo "Test Summary"
echo "================================="
echo "Total Test Suites: $TOTAL_TESTS"
echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
if [ $FAILED_TESTS -gt 0 ]; then
    echo -e "${RED}Failed: $FAILED_TESTS${NC}"
else
    echo "Failed: $FAILED_TESTS"
fi
echo

# Coverage report
if [ -n "$COVERAGE" ] && [ -f "coverage.out" ]; then
    echo "================================="
    echo "Coverage Report"
    echo "================================="
    go tool cover -func=coverage.out | tail -n 1
    echo
    echo "Generate HTML coverage report with:"
    echo "  go tool cover -html=coverage.out -o coverage.html"
    echo
fi

# Exit with appropriate status
if [ $FAILED_TESTS -gt 0 ]; then
    echo -e "${RED}Some tests failed${NC}"
    exit 1
else
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
fi
