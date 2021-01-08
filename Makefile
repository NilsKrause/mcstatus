APP=mcstatus

run: build
	./$(APP)

doc:
	godoc -http=:6090 -index

build:
	go build

debs:
	go get ./...

update-debs:
	go get -u ./...

fmt:
	gofmt -l -s .

clean:
	go clean

proto:
	protoc --go_out=/home/michael/go/src protobuf/*.proto

.PHONY: run doc build debs update-debs fmt proto