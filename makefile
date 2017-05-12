CC=go

all: TapestryD TapestryNetworking

TapestryD: TapestryD.go
	$(CC) build TapestryD.go

TapestryNetworking: TapestryNetworking.go
	$(CC) build TapestryNetworking.go

