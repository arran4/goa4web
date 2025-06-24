# This Dockerfile is used by GoReleaser to build container images.
# Keep the build stage in sync with goreleaser.yaml.
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY . .
RUN go build -o /goa4web ./cmd/goa4web

FROM scratch
COPY --from=build /goa4web /goa4web
# Image uploads are stored under /data/imagebbs inside the container.
# Future GoReleaser versions may build this image directly from the Dockerfile.
VOLUME ["/data/imagebbs"]
ENTRYPOINT ["/goa4web"]
