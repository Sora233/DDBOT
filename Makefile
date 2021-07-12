BUILD_TIME := $(shell date --rfc-3339=seconds)
COMMIT_ID := $(shell git rev-parse HEAD)

LDFLAGS = -X "main.BuildTime='"$(BUILD_TIME)"'" -X "main.CommitId='"$(COMMIT_ID)"'"

SRC := $(shell find . -type f -name '*.go')
COV := .coverage.out
TARGET := DDBOT

$(COV): $(SRC)
	go test ./... -tags=nocv -p=1 -coverprofile=$(COV)


$(TARGET): $(SRC) go.mod go.sum
ifdef NOCV
	echo 'build without opencv'
	go build -tags nocv -ldflags '$(LDFLAGS)' -o $(TARGET)
else
	echo 'build with opencv'
	go build -ldflags '$(LDFLAGS)' -o $(TARGET)
endif

build: $(TARGET)

test: $(COV)

coverage: $(COV)
	go tool cover -func=$(COV) | grep -v 'pb.go'

report: $(COV)
	go tool cover -html=$(COV)

clean:
	- rm -rf $(TARGET) $(COV)
