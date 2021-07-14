OUT=./pagr
SRC=pagr.go config.go content.go template.go

all:
	go clean
	go build -o ${OUT} ${SRC}
	go test -v
