BUILD_TIME := $(shell date --rfc-3339=seconds)
COMMIT_ID := $(shell git rev-parse HEAD)

LDFLAGS = -X "main.BuildTime='"$(BUILD_TIME)"'" -X "main.CommitId='"$(COMMIT_ID)"'"

SRC := $(shell find . -type f -name '*.go')
COV := .coverage.out
TARGET := DDBOT

$(COV): $(SRC)
	go test ./... -coverprofile=$(COV)


$(TARGET): $(SRC) go.mod go.sum
	go build -ldflags '$(LDFLAGS)' -o $(TARGET) github.com/Sora233/DDBOT/cmd

build: $(TARGET)

test: $(COV)

coverage: $(COV)
	go tool cover -func=$(COV) | grep -v 'pb.go'

report: $(COV)
	go tool cover -html=$(COV)

clean:
	- rm -rf $(TARGET) $(COV)
