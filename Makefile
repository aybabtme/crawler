
COVERPROFILE=/tmp/c.out

all:

# http://cloc.sourceforge.net/
cloc:
	@cloc --not-match-f='Makefile|_test.go' .

cover: fmt
	go test . -coverprofile=$(COVERPROFILE)
	go tool cover -html=$(COVERPROFILE)
	rm $(COVERPROFILE)

# go get github.com/kisielk/errcheck
errcheck:
	@echo "=== errcheck ==="
	@errcheck ./...

vet:
	@echo "==== go vet ==="
	@go vet ./...

lint:
	@echo "==== go lint ==="
	@golint **.go

fmt:
	@go fmt ./...

test: fmt vet lint errcheck
	@echo "=== TESTS ==="
	@godep go test ./... -parallel 8 -cpu 1,2,4 -cover
	@echo ""
	@echo ""
	@echo "=== RACE DETECTOR ==="
	@godep go test -race


.PHONY: cloc cover lint fmt test watch
