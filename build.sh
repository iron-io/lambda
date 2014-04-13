
# builds for each OS

GOOS=windows  GOARCH=amd64 go build -o bin/win/ironcli.exe
GOOS=linux    GOARCH=amd64 go build -o bin/linux/ironcli
GOOS=darwin   GOARCH=amd64 go build -o bin/mac/ironcli
