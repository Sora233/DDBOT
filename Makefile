SRC := $(shell find . -type f -name '*.go')
COV := .coverage.out
TARGET := DDBOT

$(COV): $(SRC)
	go test ./... -tags=nocv -coverprofile=$(COV)


$(TARGET): $(SRC)
	go build -o $(TARGET)

build: $(TARGET)

test: $(COV)

coverage: $(COV)
	go tool cover -func=$(COV)

clean:
	- rm -rf $(TARGET) $(COV)
