test:
	go test --cover -covermode=count -coverprofile=coverage.out ./...

build:
	make linux
	make windows
	make osx

build-tag:
	make build

linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/s4 ./main.go
	cd ./bin && tar -czf s4_linux_x64.tar.gz s4
	rm ./bin/s4

windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ./bin/s4.exe ./main.go
	cd ./bin && tar -czf s4_win_x64.tar.gz s4.exe
	rm ./bin/s4.exe

osx:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ./bin/s4 ./main.go
	cd ./bin && tar -czf s4_osx_x64.tar.gz s4
	rm ./bin/s4