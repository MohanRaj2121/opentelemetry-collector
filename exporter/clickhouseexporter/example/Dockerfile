FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM golang:latest

ARG USER_UID=10001
USER ${USER_UID}

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY ../../../cmd/otelcorecol/local /otelcorecol
ENTRYPOINT ["/otelcorecol"]
EXPOSE 4317 55680 55679
