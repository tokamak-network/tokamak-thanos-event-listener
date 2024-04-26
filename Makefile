BUILD_DIR=./build
TARGET=$(BUILD_DIR)/thanos-app-notif

SOURCE_FILES := $(shell find . -name *.go)

.PHONY: all run clean

all: build

build: $(TARGET)

run: $(TARGET)
	$<

$(TARGET): $(SOURCE_FILES)
	go build -o $@ ./cmd/thanosnotif/main.go
	@echo "Done building"

image:
	docker build --build-arg TARGETARCH=amd64 --build-arg TARGETOS=linux -t thanos-app-notif .

clean:
	rm -rf $(BUILD_DIR)