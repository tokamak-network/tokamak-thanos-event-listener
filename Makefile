BUILD_DIR=./build
TARGET=$(BUILD_DIR)/app-notif

SOURCE_FILES := $(shell find . -name *.go)

.PHONY: all run clean

all: build

build: $(TARGET)

run: $(TARGET)
	$<

$(TARGET): $(SOURCE_FILES)
	go build -o $@ ./cmd/notif/main.go
	@echo "Done building"

clean:
	rm -rf $(BUILD_DIR)