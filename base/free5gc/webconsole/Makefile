GO_BIN_PATH = bin

WEBCONSOLE = webconsole

WEBCONSOLE_GO_FILES = $(shell find -name "*.go" ! -name "*_test.go")
WEBCONSOLE_JS_FILES = $(shell find ./frontend -name '*.tsx' ! -path "*/node_modules/*")
WEBCONSOLE_FRONTEND = ./public

debug: GCFLAGS += -N -l

$(WEBCONSOLE): $(GO_BIN_PATH)/$(WEBCONSOLE) $(WEBCONSOLE_FRONTEND)

$(GO_BIN_PATH)/$(WEBCONSOLE): server.go $(WEBCONSOLE_GO_FILES)
	@echo "Start building $(@F)...."
	CGO_ENABLED=0 go build -ldflags "$(WEBCONSOLE_LDFLAGS)" -o $@ ./server.go

$(WEBCONSOLE_FRONTEND): $(WEBCONSOLE_JS_FILES)
	@echo "Start building $(@F) frontend...."
	cd frontend && \
	sudo corepack enable && \
	yarn install && \
	yarn build && \
	rm -rf ../public && \
	cp -R build ../public

clean:
	rm -rf $(GO_BIN_PATH)/$(WEBCONSOLE)
	rm -rf public