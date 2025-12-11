# Build the goa4web binary and package it into a minimal image.
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY . .
RUN go build -tags=ses -o /goa4web ./cmd/goa4web

FROM scratch
# Install the application into the final image.
ENV PATH=/usr/local/bin
ENV AUTO_MIGRATE=false
COPY --from=build /goa4web /usr/local/bin/goa4web
# Image uploads are stored under /data/imagebbs inside the container.
VOLUME ["/data/imagebbs"]
ENTRYPOINT ["/usr/local/bin/goa4web"]
CMD ["serve"]
