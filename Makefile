
testhtml:
	go test github.com/mihaichiorean/monidog/... -coverprofile=cover.out && go tool cover -html=cover.out -o coverage.html

