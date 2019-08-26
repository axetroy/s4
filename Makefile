test:
	go test --cover -covermode=count -coverprofile=coverage.out ./...

build:
	make linux
	make windows
	make macos

linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -o ./bin/linux_x86_s4 ./main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/linux_x64_s4 ./main.go

windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -o ./bin/windows_x86_s4.exe ./main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ./bin/windows_x64_s4.exe ./main.go

macos:
	CGO_ENABLED=0 GOOS=darwin GOARCH=386 go build -o ./bin/osx_x86_s4 ./main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ./bin/osx_x64_s4 ./main.go