BINARY_NAME=scotchApp

build:
	@go mod vendor
	@echo "Building Scotch..."
	@go build -o tmp/${BINARY_NAME} .
	@echo "Scotch built"

run: build
	@echo "Starting Scotch..."
	@./tmp/${BINARY_NAME} &
	@echo "Scotch Started"

clean:
	@echo "Cleaning..."
	@go clean
	@rm  tmp/${BINARY_NAME}
	@echo "Cleaned !!"

test:
	@echo "Testing..."
	@go test ./..
	@echo "Done!"

start: run


stop:
	@echo "Stopping Scotch"
	@-pkill -SIGTERM -f "./tmp/${BINARY_NAME}"
	@echo "Stopped Scotch!!"

restart: stop start 

start_compose:
	docker compose up -d

stop_compose:
	docker compose down

## test: run all tests
test:
	@go test -v ./...

## cover: opens coverage in browser
cover:
	@go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out


## coverage: displays test coverage
coverage:
	@go test -cover ./...
	

## test: run all tests
test_integ:
	@go test -cover ./... --tags integration --count=1

## cover: opens coverage in browser
cover_integ:
	@go test -coverprofile=coverage.out ./... --tags integration && go tool cover -html=coverage.out


## coverage: displays test coverage
coverage_integ:
	@go test -cover ./... --tags integration