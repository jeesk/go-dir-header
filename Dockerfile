FROM busybox:latest

COPY /go-dir-header /go-dir-header

ENTRYPOINT [ "/go-dir-header" ]
