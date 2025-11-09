.PHONY: all

all: server

server: $(wildcard *.go)
	go build -o server $^ 
