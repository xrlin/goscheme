SET CGO_ENABLED=0
SET GOOS=darwin
SET GOARCH=amd64
go build -o bin/goscheme-darwin-64 cmd/goscheme/main.go

SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
go build -o bin/goscheme-linux-64 cmd/goscheme/main.go

SET CGO_ENABLED=0
SET GOOS=windows
SET GOARCH=amd64
go build -o bin/goscheme-64.exe cmd/goscheme/main.go
