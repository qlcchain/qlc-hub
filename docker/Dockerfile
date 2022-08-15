# Build gqlc in a stock Go builder container
FROM golang:1.16.2-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers git

COPY . /qlcchain/qlc-hub
RUN cd /qlcchain/qlc-hub && make clean build

# Pull gqlc into a second stage deploy alpine container
FROM alpine:3.16.2
LABEL maintainer="developers@qlink.mobi"

ENV HUBHOME /qlchub

RUN apk --no-cache add ca-certificates && \
    addgroup qlchub && \
    adduser -S -G qlchub qlchub -s /bin/sh -h "$HUBHOME" && \
    chown -R qlchub:qlchub "$HUBHOME"

USER qlchub

WORKDIR $HUBHOME

COPY --from=builder /qlcchain/qlc-hub/build/ghub /usr/local/bin/ghub

ENTRYPOINT [ "ghub"]

VOLUME [ "$HUBHOME" ]
