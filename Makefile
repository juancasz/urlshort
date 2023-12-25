run_with_file:
	go run cmd/file/main.go -yaml=$(YAML) -json=$(JSON)

run_with_redis:
	go run cmd/redis/main.go

tests:
	go test ./... -coverprofile cover.out && go tool cover -func cover.out