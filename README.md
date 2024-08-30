smlsynctodede

# run
go run cmd/main.go

# build windows 64
GOOS=windows GOARCH=amd64 go build -o build/smlsynctodede.exe cmd/main.go