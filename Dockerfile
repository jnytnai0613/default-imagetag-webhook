FROM golang:1.18 AS build-stage

WORKDIR /kubewebhook
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go

# Build
RUN CGO_ENABLED=0 go build -o /bin/main main.go

# Final image.
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=build-stage /bin/main /usr/local/bin/main
ENTRYPOINT ["/usr/local/bin/main"]
