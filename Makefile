.RECIPEPREFIX := >
.PHONY: fmt fmt-check lint quality test security check

fmt:
>gofmt -w *.go

fmt-check:
>@test -z "$(gofmt -l .)" || (echo "Run 'make fmt' to format files" && gofmt -l . && exit 1)

lint:
>go vet ./...

quality:
>go test -race ./...

test:
>go test ./...

security:
>go install github.com/securego/gosec/v2/cmd/gosec@v2.26.1
>$(shell go env GOPATH)/bin/gosec ./...

check: fmt-check lint test security
