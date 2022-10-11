FROM golang:1.18-alpine

ENV CODE_DIR /go/src/integration_tests
ENV CGO_ENABLED 0
ENV TEST_DB_DSN=

WORKDIR ${CODE_DIR}

RUN go install github.com/onsi/ginkgo/v2/ginkgo@latest

CMD ginkgo version
