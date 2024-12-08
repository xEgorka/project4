tidy:
	go mod tidy

mock:
	mockgen -destination=internal/app/mocks/mock_storage.go -package=mocks github.com/xEgorka/project4/internal/app/storage Storage

test:
	go test ./... -coverprofile ./coverage.out
	go tool cover -func ./coverage.out
	go tool cover -html=./coverage.out -o ./coverage.html

swag:
	swag init -d cmd/ -d internal/ -g app/server/server.go  --output ./swagger/

run:
	go run -ldflags "-X github.com/xEgorka/project4/internal/app/server.buildVersion=v0.1 -X 'github.com/xEgorka/project4/internal/app/server.buildDate=$(shell date +'%Y-%m-%d')' -X github.com/xEgorka/project4/internal/app/server.buildCommit=$(shell git rev-parse --short HEAD)" cmd/main.go
