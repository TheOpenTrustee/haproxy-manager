#!/bin/ash

wait_file() {
  local file="$1"; shift
  local wait_seconds="${1:-10}"; shift # 10 seconds as default timeout
  test $wait_seconds -lt 1 && echo 'At least 1 second is required' && return 1

  until test $((wait_seconds--)) -eq 0 -o -e "$file" ; do sleep 1; done

  test $wait_seconds -ge 0 # equivalent: let ++wait_seconds
}

wait_file "/run/haproxy.pid" && {
  echo "HAProxy pid file found"
}

yq -i ".dataplaneapi.advertised.api_port = $HAPROXY_DATAPLANE_DEFAULT_PORT" /etc/haproxy/dataplaneapi.yaml

/usr/bin/dataplaneapi --host 0.0.0.0 \
    --port $HAPROXY_DATAPLANE_DEFAULT_PORT \
    --haproxy-bin /usr/local/sbin/haproxy \
    --config-file /usr/local/etc/haproxy/haproxy.cfg \
    --reload-cmd "kill -SIGUSR2 \$(cat /run/haproxy.pid)" \
    --restart-cmd "kill -SIGUSR2 \$(cat /run/haproxy.pid)" \
    --reload-delay 5 \
    --userlist $HAPROXY_DATAPLANE_USERLIST \
    --log-level trace

cat /etc/haproxy/dataplaneapi.yaml