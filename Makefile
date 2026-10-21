GO ?= go

.PHONY: fmt fmt-check test race vet check examples

fmt:
	$(GO) fmt ./...

fmt-check:
	@test -z "$$(gofmt -l $$(git ls-files '*.go'))" || (echo "Go files need gofmt" && exit 1)

test:
	$(GO) test ./...

race:
	$(GO) test -race ./...

vet:
	$(GO) vet ./...

examples:
	$(GO) run ./cmd/sml validate examples/quickstart.sml
	$(GO) run ./cmd/sml validate examples/assets
	$(GO) run ./cmd/sml render -set app.health=online -set app.cpu=37 examples/quickstart.sml
	$(GO) run ./examples/host-actions

check: fmt-check test vet examples
