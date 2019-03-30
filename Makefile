REGISTRY_NAME = docker.io/sbezverk
IMAGE_VERSION = 0.0.0

.PHONY: all syslog2cloud container push clean test

ifdef V
TESTARGS = -v -args -alsologtostderr -v 5
else
TESTARGS =
endif

all: syslog2cloud

syslog2cloud:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o ./bin/syslog2cloud ./cmd/syslog2cloud

mac-syslog2cloud:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=darwin go build -a -ldflags '-extldflags "-static"' -o ./bin/syslog2cloud.mac ./cmd/syslog2cloud

container: syslog2cloud client server
	docker build -t $(REGISTRY_NAME)/syslog2cloud:$(IMAGE_VERSION) -f ./build/Dockerfile.syslog2cloud .

push: container
	docker push $(REGISTRY_NAME)/syslog2cloud:$(IMAGE_VERSION)

clean:
	rm -rf bin

test:
	go test `go list ./... | grep -v 'vendor'` $(TESTARGS)
	go vet `go list ./... | grep -v vendor`
