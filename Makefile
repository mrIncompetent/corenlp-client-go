BUF_VERSION=v0.41.0
PROTOC_GEN_GO_VERSION=v1.26.0
CORENLP_VERSION=v4.2.1
BIN_PATH=./bin
BUF=$(BIN_PATH)/buf
PROTOC_GEN_GO=$(BIN_PATH)/protoc-gen-go
GOLANGCI_LINT=$(BIN_PATH)/golangci-lint
GOLANGCI_LINT_VERSION=v1.40.0

$(BIN_PATH):
	mkdir $(BIN_PATH)

$(GOLANGCI_LINT):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(BIN_PATH) $(GOLANGCI_LINT_VERSION)

$(PROTOC_GEN_GO):
	GOBIN=$(realpath $(BIN_PATH)) go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)

$(BUF):
	curl -L -o $(BUF) https://github.com/bufbuild/buf/releases/download/$(BUF_VERSION)/buf-Linux-x86_64
	chmod +x $(BUF)

install_gen_deps: $(BIN_PATH) $(BUF) $(PROTOC_GEN_GO)

gen: install_gen_deps
	$(BUF) generate --template buf.gen.yaml "https://github.com/stanfordnlp/CoreNLP.git#tag=$(CORENLP_VERSION)" --path src/edu/stanford/nlp/pipeline/

.PHONY: lint
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run --config ./.golangci.yml
