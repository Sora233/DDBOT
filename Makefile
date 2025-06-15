BUILD_TIME := $(shell date --rfc-3339=seconds)
COMMIT_ID := $(shell git rev-parse HEAD)

LDFLAGS = -X "github.com/Sora233/DDBOT/v2/lsp.BuildTime='"$(BUILD_TIME)"'" -X "github.com/Sora233/DDBOT/v2/lsp.CommitId='"$(COMMIT_ID)"'"

SRC := $(shell find . -type f -name '*.go') lsp/template/default/*
PROTO := $(shell find . -type f -name '*.proto')
COV := .coverage.out
TARGET := DDBOT

$(COV): $(SRC)
	go test ./... -coverprofile=$(COV)


$(TARGET): $(SRC) go.mod go.sum
	go build -ldflags '$(LDFLAGS)' -o $(TARGET) github.com/Sora233/DDBOT/v2/cmd

build: $(TARGET)

proto: $(PROTO)
	protoc --go_out=. $(PROTO)

test: $(COV)

coverage: $(COV)
	go tool cover -func=$(COV) | grep -v 'pb.go'

report: $(COV)
	go tool cover -html=$(COV)

clean:
	- rm -rf $(TARGET) $(COV)
