SRC := $(shell find . -type f -name '*.go')
COV := .coverage.out
TARGET := DDBOT

$(COV): $(SRC)
	go test ./... -tags=nocv -p=1 -coverprofile=$(COV)


$(TARGET): $(SRC)
	go build -o $(TARGET)

build: $(TARGET)

test: $(COV)

coverage: $(COV)
	go tool cover -func=$(COV)

report: $(COV)
	go tool cover -html=$(COV)

clean:
	- rm -rf $(TARGET) $(COV)
