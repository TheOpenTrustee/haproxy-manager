FROM golang:1.24-alpine AS gobuild
WORKDIR /app

# Copy go package for building
COPY . .

RUN apk add --no-cache musl-dev gcc
RUN go mod tidy && CGO_ENABLED=1 go build ./cmd/webapp

FROM haproxytech/haproxy-alpine:3.1
WORKDIR /app


COPY --from=gobuild /app/webapp ./

COPY scripts/_utils.sh ./
COPY scripts/manager.sh ./


# Install dependencies
ENV PROCFUSTION_VERSION=v0.2.2
RUN wget -O procfusion.tar.gz https://github.com/linkdd/procfusion/releases/download/${PROCFUSTION_VERSION}/procfusion-${PROCFUSTION_VERSION}-x86_64-unknown-linux-musl.tar.gz && \
    tar -xvf procfusion.tar.gz --strip-components=1 && \
    rm procfusion.tar.gz LICENSE.txt README.md && \
    apk add yq

# haproxy dataplane settings
ENV HAPROXY_DATAPLANE_USER=admin
ENV HAPROXY_DATAPLANE_PASS=mypassword
ENV HAPROXY_DATAPLANE_DEFAULT_PORT=5555
ENV HAPROXY_DATAPLANE_USERLIST=default-haproxy-dataplane
COPY scripts/dataplane.sh ./dataplane.sh
COPY configs/dataplaneapi.yaml /etc/haproxy/dataplaneapi.yaml

# haproxy settings
COPY configs/haproxy.cfg /usr/local/etc/haproxy/haproxy.cfg

# procfusion used for process management because haproxy is deprecating "program"
ENV PROCFUSION_CONFIG=config.toml
COPY scripts/procfusion.sh ./
COPY configs/config.toml ./

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["./procfusion.sh"]