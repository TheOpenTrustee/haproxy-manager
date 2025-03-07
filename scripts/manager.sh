#!/bin/ash
SCRIPT_DIR=$(dirname "$0")
source "$SCRIPT_DIR/_utils.sh"

wait_file "/run/haproxy.pid" && {
  echo "HAProxy pid file found"
}

# Define the endpoint URL
ENDPOINT="http://localhost:5555/v3/specification"

# Call the check_endpoint function with a 15-second timeout
check_endpoint "$ENDPOINT" 15

# Check the result
if [ $? -eq 0 ]; then
  echo "Endpoint is up, continuing..."
else
  echo "Enpoint is down, timeout reached."
fi

./test