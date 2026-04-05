build:
	CGO_ENABLED=0 go build -o checkin ./cmd/checkin/

run: build
	./checkin

test:
	go test ./...

clean:
	rm -f checkin

.PHONY: build run test clean
