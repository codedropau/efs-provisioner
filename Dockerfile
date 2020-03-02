FROM golang:1.13
ENV GO111MODULE=on
ADD . /go/src/github.com/codedropau/efs-provisioner
WORKDIR /go/src/github.com/codedropau/efs-provisioner
RUN go get github.com/mitchellh/gox
RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=0 /go/src/github.com/codedropau/efs-provisioner/bin/efs-provisioner_linux_amd64 /usr/local/bin/efs-provisioner
CMD ["efs-provisioner"]
