FROM golang:1.18 AS build

WORKDIR .

# Build with goreleaser
RUN go install github.com/goreleaser/goreleaser@latest && \
    goreleaser --snapshot --skip-publish --rm-dist

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go

# Move binary into final image
FROM gcr.io/distroless/static-debian11 AS app
COPY --from=build /dist/galax-app-linux-amd64 /galax-app
CMD ["/galax-app"]