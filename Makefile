.PHONY: all
all: x1 x1tool

GOPROXY ?= "https://proxy.golang.org,direct"
.PHONY: x1 x1tool
x1:
	GIT_COMMIT=`git rev-list -1 HEAD 2>/dev/null || echo ""` && \
	GIT_DATE=`git log -1 --date=short --pretty=format:%ct 2>/dev/null || echo ""` && \
	GOPROXY=$(GOPROXY) \
	go build \
	    -ldflags "-s -w -X github.com/Fantom-foundation/go-opera/config.GitCommit=$${GIT_COMMIT} -X github.com/Fantom-foundation/go-opera/config.GitDate=$${GIT_DATE}" \
	    -o build/x1 \
	    ./cmd/sonicd

x1tool:
	GIT_COMMIT=`git rev-list -1 HEAD 2>/dev/null || echo ""` && \
	GIT_DATE=`git log -1 --date=short --pretty=format:%ct 2>/dev/null || echo ""` && \
	GOPROXY=$(GOPROXY) \
	go build \
	    -ldflags "-s -w -X github.com/Fantom-foundation/go-opera/config.GitCommit=$${GIT_COMMIT} -X github.com/Fantom-foundation/go-opera/config.GitDate=$${GIT_DATE}" \
	    -o build/x1tool \
	    ./cmd/sonictool

TAG ?= "latest"
.PHONY: x1-image
x1-image:
	docker build \
    	    --network=host \
    	    -f ./docker/Dockerfile.opera -t "x1:$(TAG)" .

.PHONY: test
test:
	go test ./...

.PHONY: coverage
coverage:
	go test -coverprofile=cover.prof $$(go list ./... | grep -v '/gossip/contract/' | grep -v '/gossip/emitter/mock' | xargs)
	go tool cover -func cover.prof | grep -e "^total:"

.PHONY: fuzz
fuzz:
	CGO_ENABLED=1 \
	mkdir -p ./fuzzing && \
	go run github.com/dvyukov/go-fuzz/go-fuzz-build -o=./fuzzing/gossip-fuzz.zip ./gossip && \
	go run github.com/dvyukov/go-fuzz/go-fuzz -workdir=./fuzzing -bin=./fuzzing/gossip-fuzz.zip


.PHONY: clean
clean:
	rm -fr ./build/*
