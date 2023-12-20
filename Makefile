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