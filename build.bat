SET APP_DIR=build
SET GOARCH=386
go build -o %APP_DIR%\smrate.exe github.com/fpawel/smrate/cmd/smrate
