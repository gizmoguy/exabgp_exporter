FROM python:3.11-alpine AS python-env

FROM golang:alpine
ENV CGO_ENABLED 0
ENV PYTHON_VERSION 3.11
ENV EXABGP_VERSION 4.2.21
ENV GOBGP_VERSION 3.23.0
ENV S6_OVERLAY_VERSION 3.1.6.2
ENV HOME /root
ENV S6_LOGGING 1
WORKDIR /root

RUN mkdir -p /exabgp/run
RUN mkdir -p /exabgp/etc/exabgp
RUN mkdir -p /gobgp
RUN mkfifo /exabgp/run/exabgp.in
RUN mkfifo /exabgp/run/exabgp.out
RUN chmod 666 /exabgp/run/exabgp.*
RUN mkfifo /exabgp/exabgp.cmd
RUN chmod 666 /exabgp/exabgp.cmd

COPY --from=python-env /usr/local/ /usr/local/

COPY docker/files/exabgp.conf /exabgp/etc/exabgp/exabgp.conf
COPY docker/files/gobgp.yaml /gobgp/gobgp.yaml
COPY docker/files/rsyslog.conf /etc/rsyslog.conf
COPY docker/*.sh /root/

# add support packages
ADD https://github.com/osrg/gobgp/releases/download/v${GOBGP_VERSION}/gobgp_${GOBGP_VERSION}_linux_amd64.tar.gz /gobgp.tar.gz
ADD https://github.com/just-containers/s6-overlay/releases/download/v${S6_OVERLAY_VERSION}/s6-overlay-noarch.tar.xz /tmp
ADD https://github.com/just-containers/s6-overlay/releases/download/v${S6_OVERLAY_VERSION}/s6-overlay-x86_64.tar.xz /tmp

# install support packages
RUN tar xvf /gobgp.tar.gz -C /gobgp/
RUN tar -C / -Jxpf /tmp/s6-overlay-noarch.tar.xz
RUN tar -C / -Jxpf /tmp/s6-overlay-x86_64.tar.xz

RUN mkdir /etc/services.d/
RUN mkdir /etc/services.d/gobgp
RUN mkdir /etc/services.d/exabgp
RUN mkdir /etc/services.d/exabgp_exporter_good
RUN mkdir /etc/services.d/exabgp_exporter_bad
COPY docker/exabgp.sh /etc/services.d/exabgp/run
COPY docker/gobgpd.sh /etc/services.d/gobgp/run
COPY docker/exporter_valid.sh /etc/services.d/exabgp_exporter_good/run
COPY docker/exporter_invalid.sh /etc/services.d/exabgp_exporter_bad/run

RUN apk add\
    bash \
    socat \
    curl \
    git \
    musl-dev \
    linux-headers \
    gcc
    libffi \

# build binary
ADD . /src
RUN cd /src && CGO_ENABLED=0 go build -o /exabgp/exabgp_exporter -ldflags "-s" -a -tags netgo ./cmd/exabgp_exporter

ENTRYPOINT [ "/root/install-and-init.sh" ]