# Build the goa4web binary and package it into a minimal image.
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY . .
RUN go build -o /goa4web ./cmd/goa4web \
    && go build -o /goa4web-admin ./cmd/goa4web-admin

FROM scratch
# Install both the main application and the admin CLI into the final image.
ENV PATH=/usr/local/bin
COPY --from=build /goa4web /usr/local/bin/goa4web
COPY --from=build /goa4web-admin /usr/local/bin/goa4web-admin
# Image uploads are stored under /data/imagebbs inside the container.
VOLUME ["/data/imagebbs"]
ENTRYPOINT ["/usr/local/bin/goa4web"]
