test:
	go test --cover -covermode=count -coverprofile=coverage.out ./...

build:
	sh build.sh
