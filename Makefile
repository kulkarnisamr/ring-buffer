GO=go

tidy:
	$(GO) mod tidy -compat=1.20

cover:
	$(GO) test ./...

build:
	$(GO) build ring_buffer.go

benchmark:
	$(GO) test -bench=. -benchmem ./...