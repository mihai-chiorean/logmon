BIN=./monidog

$(BIN): build

build:
	go build

run: $(BIN)
	./monidog

run-integration:
	cd integration && go run main.go

run-test: $(BIN)
	./monidog --log testing/access.log

testhtml:
	go test github.com/mihaichiorean/monidog/... -coverprofile=cover.out && go tool cover -html=cover.out -o coverage.html

