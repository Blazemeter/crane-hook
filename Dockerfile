
FROM --platform=linux/amd64 golang:1.26.2-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/cranehook .

FROM --platform=linux/amd64 alpine:3.22 AS certs
RUN apk --no-cache add ca-certificates

FROM --platform=linux/amd64 scratch

WORKDIR /usr/local/bin
COPY --from=builder /out/cranehook /usr/local/bin/cranehook
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Set environment variables
ENV WORKING_NAMESPACE=""
ENV ROLE_NAME=""
ENV ROLE_BINDING_NAME=""
ENV SERVICE_ACCOUNT_NAME=""
ENV KUBERNETES_WEB_EXPOSE_TYPE=""
# ENV SV_ENABLE=""
ENV KUBERNETES_ISTIO_GATEWAY_NAME=""
ENV DOCKER_REGISTRY=""
ENV KUBERNETES_WEB_EXPOSE_TLS_SECRET_NAME=""
ENV HTTP_PROXY=""
ENV HTTPS_PROXY=""
ENV NO_PROXY=""

USER 65532:65532

CMD ["/usr/local/bin/cranehook"]
