FROM alpine:latest
ADD aws-container-factory /tmp/aws-container-factory
ENTRYPOINT ["/tmp/aws-container-factory"]