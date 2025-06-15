FROM alpine:3.18
RUN adduser -D -u 10001 app
USER app
COPY goa4web /usr/local/bin/goa4web
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/goa4web"]
