run_with_file:
	go run cmd/file/main.go -yaml=$(YAML) -json=$(JSON)

run_with_redis:
	sh scripts/run_with_redis.sh

stop_with_redis:
	sh scripts/stop_with_redis.sh

unit_tests:
	go test ./... -coverprofile cover.out && go tool cover -func cover.out

integration_tests:
	sh scripts/run_integration_tests.sh