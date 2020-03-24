.PHONY: all

all: membership.so

membership.so:
	go build -o membership.so -buildmode=plugin handler.go
