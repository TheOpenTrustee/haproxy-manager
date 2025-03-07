#!/bin/ash

wait_file() {
  local file="$1"; shift
  local wait_seconds="${1:-10}"; shift # 10 seconds as default timeout
  test $wait_seconds -lt 1 && echo 'At least 1 second is required' && return 1

  until test $((wait_seconds--)) -eq 0 -o -e "$file" ; do sleep 1; done

  test $wait_seconds -ge 0 # equivalent: let ++wait_seconds
}

check_endpoint() {
  local url="$1"; shift
  local wait_seconds="${1:-10}"; shift # 10 seconds as default timeout
  test $wait_seconds -lt 1 && echo 'At least 1 second is required' && return 1

  # Function to check if the endpoint is up
  is_endpoint_up() {
    wget --quiet --spider "$url" 2>/dev/null
  }

  until test $((wait_seconds--)) -eq 0 -o "$(is_endpoint_up)" ; do
    sleep 1
  done

  test $wait_seconds -ge 0 # Return true if endpoint is up before timeout
}