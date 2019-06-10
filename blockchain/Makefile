all: install

install: go.sum
    GO111MODULE=on go install -tags "$(build_tags)" ./cmd/kavad
    GO111MODULE=on go install -tags "$(build_tags)" ./cmd/kavacli

go.sum: go.mod
		@echo "--> Ensure dependencies have not been modified"
    GO111MODULE=on @go mod verify