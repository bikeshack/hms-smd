# MIT License
#
# (C) Copyright [2019-2023] Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.

# Dockerfile for building HMS State Manager.


### Build Base Stage ###
# Build base just has the packages installed we need.
FROM golang:1.16-alpine AS build-base

RUN set -ex \
    && apk -U upgrade \
    && apk add build-base


### Base Stage ###
# Base copies in the files we need to test/build.
FROM build-base AS base

RUN go env -w GO111MODULE=auto

# Copy all the necessary files to the image.
RUN mkdir /build/
COPY cmd /build/cmd
COPY internal /build/internal
COPY pkg /build/pkg
COPY go.mod go.sum /build/

WORKDIR /build



### Build Stage ###
FROM base AS builder

# Base image contains everything needed for Go building, just build.
RUN set -ex \
    && go mod tidy \
    && go build -o smd -v -tags musl ./cmd/smd \
    && go build -o smd-loader -v -tags musl ./cmd/smd-loader \
    && go build -o smd-init -v -tags musl ./cmd/smd-init


### Final Stage ###
FROM alpine:3.15
LABEL maintainer="Hewlett Packard Enterprise" 
EXPOSE 27779
STOPSIGNAL SIGTERM

ENV SMD_DBNAME="hmsds"
ENV SMD_DBUSER="hmsdsuser"
ENV SMD_DBTYPE="postgres"
ENV SMD_DBPORT=5432
ENV SMD_DBOPTS="sslmode=disable"

# yaml should always overwrite password using kubernetes secrets in production.
ENV SMD_DBHOST="cray-smd-postgres"
ENV SMD_DBPASS="hmsdsuser"

ENV TLSCERT=""
ENV TLSKEY=""

ENV LOGLEVEL 2

ENV SMD_HWINVHIST_AGE_MAX_DAYS=365

ENV SMD_CA_URI=""

# Copy the final binary
COPY --from=builder /build/smd /usr/local/bin
COPY --from=builder /build/smd-loader /usr/local/bin
COPY --from=builder /build/smd-init /usr/local/bin

COPY configs /configs

WORKDIR /

# Cannot live without these packages installed.
RUN set -ex \
    && apk -U upgrade \
    && apk add --no-cache \
        postgresql-client \
    && mkdir -p /persistent_migrations \
    && chmod 777 /persistent_migrations

COPY migrations/postgres /persistent_migrations

# nobody 65534:65534
USER 65534:65534

CMD ["sh", "-c", "smd -db-dsn=$DBDSN -tls-cert=$TLSCERT -tls-key=$TLSKEY -log=$LOGLEVEL -dbhost=$POSTGRES_HOST -dbport=$POSTGRES_PORT"]
