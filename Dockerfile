# Build the goa4web binary and package it into a minimal image.
FROM golang:1.22-alpine AS build
RUN apk add --no-cache ca-certificates
WORKDIR /src
COPY . .
RUN go build -tags=ses -o /goa4web ./cmd/goa4web

FROM alpine:3.20 AS runtime
RUN addgroup -S goa4web && adduser -S -G goa4web -u 65532 goa4web \
  && mkdir -p /data/imagebbs \
  && chown -R goa4web:goa4web /data

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
# Install the application into the final image.
ENV PATH=/usr/local/bin
ENV AUTO_MIGRATE=false
COPY --from=runtime /etc/passwd /etc/passwd
COPY --from=runtime /etc/group /etc/group
COPY --from=runtime /data /data
COPY --from=build /goa4web /usr/local/bin/goa4web
# Image uploads are stored under /data/imagebbs inside the container.
VOLUME ["/data/imagebbs"]
USER goa4web
ENTRYPOINT ["/usr/local/bin/goa4web"]
CMD ["serve"]
