.PHONY: env test golangci-lint gosec govulncheck verify

GOPROXY := https://goproxy.cn,direct
export GOPROXY
GO_LINT=$(shell which golangci-lint 2> /dev/null || echo '')
GO_LINT_URI=github.com/golangci/golangci-lint/cmd/golangci-lint@latest
GO_SEC=$(shell which gosec 2> /dev/null || echo '')
GO_SEC_URI=github.com/securego/gosec/v2/cmd/gosec@latest
GO_VULNCHECK=$(shell which govulncheck 2> /dev/null || echo '')
GO_VULNCHECK_URI=golang.org/x/vuln/cmd/govulncheck@latest

default: test

env:
	@go version
test: env
	go test -race -cover -coverpkg ./... -coverprofile=coverage -covermode=atomic ./...
	@go tool cover -html=coverage -o coverage.html
	@go tool cover -func=coverage -o coverage.txt
	@tail -n 1 coverage.txt

golangci-lint:
	$(if $(GO_LINT), ,go install $(GO_LINT_URI))
	@echo "##### Running golangci-lint"
	golangci-lint run -D staticcheck -D unused --timeout=2m

gosec:
	$(if $(GO_SEC), ,go install $(GO_SEC_URI))
	@echo "##### Running gosec"
	gosec ./...

govulncheck:
	$(if $(GO_VULNCHECK), ,go install $(GO_VULNCHECK_URI))
	@echo "##### Running govulncheck"
	govulncheck ./...

verify: golangci-lint gosec
