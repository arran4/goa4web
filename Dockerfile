FROM scratch
COPY goa4web /goa4web
# Image uploads are stored under /data/imagebbs inside the container.
VOLUME ["/data/imagebbs"]
ENTRYPOINT ["/goa4web"]
