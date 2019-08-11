go fmt
./runTests.sh
GOOS=linux GOARCH=amd64 go build
GOOS=windows GOARCH=amd64 go build
