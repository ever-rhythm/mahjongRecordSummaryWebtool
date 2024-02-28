go env -w CGO_ENABLED=0
go env -w GOOS=linux
go env -w GOARCH=amd64
go build -o ./t20240226_v2 -ldflags="-s -w" main.go