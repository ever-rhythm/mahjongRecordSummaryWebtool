go env -w CGO_ENABLED=0
go env -w GOOS=linux
go env -w GOARCH=amd64
go build -o ./t20240220_v4 -ldflags="-s -w" main.go