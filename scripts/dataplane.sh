#!/bin/ash

SCRIPT_DIR=$(dirname "$0")
source "$SCRIPT_DIR/_utils.sh"

wait_file "/run/haproxy.pid" && {
  echo "HAProxy pid file found"
}

yq -i ".dataplaneapi.advertised.api_port = $HAPROXY_DATAPLANE_DEFAULT_PORT" /etc/haproxy/dataplaneapi.yaml

/usr/local/bin/dataplaneapi --host 0.0.0.0 \
    --port $HAPROXY_DATAPLANE_DEFAULT_PORT \
    --haproxy-bin /usr/local/sbin/haproxy \
    --config-file /usr/local/etc/haproxy/haproxy.cfg \
    --reload-cmd "kill -SIGUSR2 \$(cat /run/haproxy.pid)" \
    --restart-cmd "kill -SIGUSR2 \$(cat /run/haproxy.pid)" \
    --reload-delay 5 \
    --userlist $HAPROXY_DATAPLANE_USERLIST \
    --log-to stdout \
    --log-level trace

cat /etc/haproxy/dataplaneapi.yaml