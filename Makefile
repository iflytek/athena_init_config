GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) unpack

build:
	rm -rf bin
	mkdir bin
	$(GOBUILD) -v -o ./bin/initc -gcflags "-N -l -c 10" ./main.go
