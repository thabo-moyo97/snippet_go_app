#!/bin/sh
set -e

# Ensure OpenSSL is installed
if ! command -v openssl >/dev/null 2>&1; then
    echo "Installing OpenSSL..."
    apk add --no-cache openssl
fi

# Generate certificates if they don't exist
if [ ! -f "/app/tls/cert.pem" ] || [ ! -f "/app/tls/key.pem" ]; then
    echo "Generating development certificates..."
    chmod +x /app/scripts/generate-dev-certs.sh
    sh /app/scripts/generate-dev-certs.sh
fi

# Default values
DEBUG_MODE=0

# Setup Go environment
if [ -z "$GOPATH" ]; then
    export GOPATH=$HOME/go
    echo "GOPATH was not set, using default: $GOPATH"
fi

# Ensure Go binaries are in PATH
export PATH=$PATH:$GOPATH/bin

# Verify Air installation
if ! command -v air >/dev/null 2>&1; then
    echo "Air is not installed. Installing..."
    go install github.com/air-verse/air@latest
fi

# Load environment variables if .env file exists
if [ -f .env ]; then
    echo "Loading environment variables from .env file..."
    set -a
    . ./.env
    set +a
fi

# Parse command line arguments
while [ "$#" -gt 0 ]; do
    case "$1" in
        --debug)
            DEBUG_MODE=1
            shift 1
            ;;
        --help)
            echo "Usage: start-dev [OPTIONS]"
            echo "Options:"
            echo "  --debug    Start the server in debug mode"
            echo "  --help     Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

if [ $DEBUG_MODE = 1 ]; then
    echo "Starting with Air and debug support..."
    exec air -c .air.debug.toml
else
    echo "Starting with Air in normal mode..."
    exec air -c .air.toml
fi