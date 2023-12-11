run:
	go run main/main.go -yaml=$(YAML) -json=$(JSON)

tests:
	go test ./... -coverprofile cover.out && go tool cover -func cover.out