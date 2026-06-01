# Portable, OCI-standard multi-stage build.
# Produces a static binary on scratch so the image runs unchanged on any
# OCI runtime / any cloud's Kubernetes (artifact transportability).
#
#   docker build --build-arg SERVICE=billing -t ghcr.io/unboxd-cloud/billing .
#
ARG GO_VERSION=1.24

FROM golang:${GO_VERSION} AS build
ARG SERVICE
WORKDIR /src
COPY go.mod ./
COPY . .
# CGO disabled => fully static, no libc dependency => transportable.
RUN CGO_ENABLED=0 GOFLAGS=-trimpath go build -ldflags="-s -w" \
    -o /out/app ./cmd/${SERVICE}

FROM scratch
COPY --from=build /out/app /app
# Document the standard ports; overridden per service via env.
EXPOSE 8080 8081 8082 8083 8084 8085
USER 65532:65532
ENTRYPOINT ["/app"]
