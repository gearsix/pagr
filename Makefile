OUT=./pagr
SRC=pagr.go config.go page.go template.go

all:
	go clean
	go build -o ${OUT} ${SRC}

test:
	go clean
	go build -o ${OUT} ${SRC}
	go test -v
