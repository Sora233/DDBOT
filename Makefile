SRC = $(wildcard **/*.go) main.go
COV = .coverage.out
$(COV): $(SRC)
	go test ./... -v -coverprofile=$(COV)
	go tool cover -func=$(COV)

TARGET = DDBOT

$(TARGET): $(SRC)
	go build -o $(TARGET)

build: $(TARGET)

test: $(COV)

clean:
	- rm -rf $(TARGET) $(COV)
