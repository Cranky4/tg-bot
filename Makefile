CURDIR=$(shell pwd)
BINDIR=${CURDIR}/bin
GOVER=$(shell go version | perl -nle '/(go\d\S+)/; print $$1;')
MOCKGEN=${BINDIR}/mockgen_${GOVER}
SMARTIMPORTS=${BINDIR}/smartimports_${GOVER}
LINTVER=v1.49.0
LINTBIN=${BINDIR}/lint_${GOVER}_${LINTVER}
PACKAGE=gitlab.ozon.dev/cranky4/tg-bot/cmd/bot
SEEDER=gitlab.ozon.dev/cranky4/tg-bot/cmd/seeder
REPORTER=gitlab.ozon.dev/cranky4/tg-bot/cmd/reporter
TG_BOT_DB="tg_bot"
TG_BOT_DB_USER="tg_bot_user"
TG_BOT_DB_PASSWORD="secret"
TG_BOT_DB_HOST="localhost"
TG_BOT_DB_PORT="5432"

.PHONY: test-coverage


all: format build test lint

build: bindir build-bot build-reporter

build-bot:
	go build -o ${BINDIR}/bot ${PACKAGE}
build-reporter:
	go build -o ${BINDIR}/reporter ${REPORTER}

test:
	go test ./internal/...
test-coverage:
	go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out

run:
	go run ${PACKAGE} 2>&1 | tee logs/bot.log

run-seeder:
	go run ${SEEDER}

run-reporter:
	go run ${REPORTER} 2>&1 | tee logs/reporter.log

generate: install-mockgen
	${MOCKGEN} \
		-source=internal/service/messages/incoming_msg.go \
		-destination=internal/service/messages/mocks/messages_mocks.go
	${MOCKGEN} \
		-source=internal/repository/expenses.go \
		-destination=internal/repository/mocks/expenses_repo_mocks.go
	${MOCKGEN} \
		-source=internal/service/expense_processor/expense_processor.go \
		-destination=internal/service/expense_processor/mocks/expense_processor_mocks.go
	${MOCKGEN} \
		-source=internal/service/report_requester/report_requester.go \
		-destination=internal/service/report_requester/mocks/report_requester_mocks.go
	${MOCKGEN} \
		-source=internal/service/expense_reporter/expense_reporter.go \
		-destination=internal/service/expense_reporter/mocks/expense_reporter_mocks.go
	${MOCKGEN} \
		-source=internal/service/cache/cache.go \
		-destination=internal/service/cache/mocks/cache_mocks.go
	${MOCKGEN} \
		-source=internal/clients/message_broker/client.go \
		-destination=internal/clients/message_broker/mocks/client_mocks.go
	${MOCKGEN} \
		-source=internal/service/report_sender/report_sender.go \
		-destination=internal/service/report_sender/mocks/report_sender_mocks.go

lint: install-lint
	${LINTBIN} run

precommit: format build test lint
	echo "OK"

bindir:
	mkdir -p ${BINDIR}

format: install-smartimports
	${SMARTIMPORTS} -exclude internal/mocks

install-mockgen: bindir
	test -f ${MOCKGEN} || \
		(GOBIN=${BINDIR} go install github.com/golang/mock/mockgen@v1.6.0 && \
		mv ${BINDIR}/mockgen ${MOCKGEN})

install-lint: bindir
	test -f ${LINTBIN} || \
		(GOBIN=${BINDIR} go install github.com/golangci/golangci-lint/cmd/golangci-lint@${LINTVER} && \
		mv ${BINDIR}/golangci-lint ${LINTBIN})

install-smartimports: bindir
	test -f ${SMARTIMPORTS} || \
		(GOBIN=${BINDIR} go install github.com/pav5000/smartimports/cmd/smartimports@latest && \
		mv ${BINDIR}/smartimports ${SMARTIMPORTS})

build-dev:
	docker compose -f deployments/docker-compose.dev.yml pull
	docker compose -f deployments/docker-compose.dev.yml build
up-dev:
	docker compose -f deployments/docker-compose.dev.yml up -d
down-dev:
	docker compose -f deployments/docker-compose.dev.yml down --remove-orphans
rest-dev: down-dev up-dev

install-goose:
	(which goose > /dev/null) || go install github.com/pressly/goose/v3/cmd/goose@latest
migrate-status: install-goose
	 goose -dir ./migrations postgres "host=${TG_BOT_DB_HOST} user=${TG_BOT_DB_USER} password=${TG_BOT_DB_PASSWORD} dbname=${TG_BOT_DB} port=${TG_BOT_DB_PORT} sslmode=disable" status

migrate-create: install-goose
	 goose -dir ./migrations postgres "host=${TG_BOT_DB_HOST} user=${TG_BOT_DB_USER} password=${TG_BOT_DB_PASSWORD} dbname=${TG_BOT_DB} port=${TG_BOT_DB_PORT} sslmode=disable" create tg_bot sql

migrate: install-goose
	 goose -dir ./migrations postgres "host=${TG_BOT_DB_HOST} user=${TG_BOT_DB_USER} password=${TG_BOT_DB_PASSWORD} dbname=${TG_BOT_DB} port=${TG_BOT_DB_PORT} sslmode=disable" up

migrate-down: install-goose	
	 goose -dir ./migrations postgres "host=${TG_BOT_DB_HOST} user=${TG_BOT_DB_USER} password=${TG_BOT_DB_PASSWORD} dbname=${TG_BOT_DB} port=${TG_BOT_DB_PORT} sslmode=disable" down

integration-tests: 
	docker-compose -f deployments/docker-compose.test.yml up -d --build
	echo "optimistic waiting for docker ready" && sleep 2
	docker-compose -f deployments/docker-compose.test.yml run tester ginkgo || docker-compose -f deployments/docker-compose.test.yml down --remove-orphans
	docker-compose -f deployments/docker-compose.test.yml down --remove-orphans

install-ginkgo:
	go get github.com/onsi/ginkgo/v2/ginkgo
	go get github.com/onsi/gomega/...

install-protobuf:
	go install google.golang.org/protobuf/cmd/protoc-gen-go
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc

gf:
	protoc ./api/Reporter.proto \
		--grpc-gateway_out ./pkg \
		--grpc-gateway_opt logtostderr=true \
		--grpc-gateway_opt paths=source_relative \
		--grpc-gateway_opt generate_unbound_methods=true

generate-grpc: 
	mkdir -p pkg/reporter_v1
	protoc --proto_path api/ \
		--go_out=pkg/reporter_v1 --go_opt=paths=import \
		--go-grpc_out=pkg/reporter_v1 --go-grpc_opt=paths=import \
		--grpc-gateway_out ./pkg/reporter_v1 \
		--grpc-gateway_opt logtostderr=true \
		--grpc-gateway_opt paths=source_relative \
		--grpc-gateway_opt generate_unbound_methods=true \
		--validate_out lang=go:pkg/reporter_v1 \
		--openapiv2_out ./pkg/reporter_v1 \
		--openapiv2_opt logtostderr=true \
		--openapiv2_opt generate_unbound_methods=true \
		api/Reporter.proto
	mv pkg/reporter_v1/gitlab.ozon.dev/cranky4/tg-bot/api/* pkg/reporter_v1
	rm -rf pkg/reporter_v1/gitlab.ozon.dev