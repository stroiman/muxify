muxify: *.go
	go build

.PHONY: build
build: muxify

.PHONY: install
install: build
	go install

.PHONY: test
test: build
	./muxify project-1
