FROM python:3.11-alpine AS python-env

FROM golang:alpine

ENV CGO_ENABLED=0
ENV PYTHON_VERSION=3.11
ENV EXABGP_VERSION=4.2.25
ENV GOBGP_VERSION=4.0.0
ENV S6_OVERLAY_VERSION=3.2.0.2
ENV HOME=/root
ENV S6_LOGGING=1

WORKDIR /root

COPY --from=python-env /usr/local/ /usr/local/

COPY docker/files/exabgp.conf /exabgp/etc/exabgp/exabgp.conf
COPY docker/files/gobgp.yaml /gobgp/gobgp.yaml
COPY docker/files/rsyslog.conf /etc/rsyslog.conf
COPY docker/*.sh /root/
COPY docker/exabgp.sh /etc/services.d/exabgp/run
COPY docker/gobgpd.sh /etc/services.d/gobgp/run
COPY docker/exporter_valid.sh /etc/services.d/exabgp_exporter_good/run
COPY docker/exporter_invalid.sh /etc/services.d/exabgp_exporter_bad/run

# add exabgp_exporter source code
ADD . /src

# add support packages
ADD https://github.com/osrg/gobgp/releases/download/v${GOBGP_VERSION}/gobgp_${GOBGP_VERSION}_linux_amd64.tar.gz /gobgp.tar.gz
ADD https://github.com/just-containers/s6-overlay/releases/download/v${S6_OVERLAY_VERSION}/s6-overlay-noarch.tar.xz /tmp
ADD https://github.com/just-containers/s6-overlay/releases/download/v${S6_OVERLAY_VERSION}/s6-overlay-x86_64.tar.xz /tmp

# install dependencies
RUN apk add \
    bash \
    socat \
    curl \
    git \
    musl-dev \
    linux-headers \
    gcc \
    libffi \
# create exabgp dirs/files \
 && mkdir -p /exabgp/run \
 && mkdir -p /exabgp/etc/exabgp \
 && mkdir -p /gobgp \
 && mkfifo /exabgp/run/exabgp.in \
 && mkfifo /exabgp/run/exabgp.out \
 && chmod 666 /exabgp/run/exabgp.* \
 && mkfifo /exabgp/exabgp.cmd \
 && chmod 666 /exabgp/exabgp.cmd \
# install support packages \
 && tar xvf /gobgp.tar.gz -C /gobgp/ \
 && tar -C / -Jxpf /tmp/s6-overlay-noarch.tar.xz \
 && tar -C / -Jxpf /tmp/s6-overlay-x86_64.tar.xz \
# build exabgp_exporter binary \
 && cd /src \
 && CGO_ENABLED=0 go build -o /exabgp/exabgp_exporter -ldflags "-s" -a -tags netgo ./cmd/exabgp_exporter

ENTRYPOINT [ "/root/install-and-init.sh" ]
