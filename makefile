CC=go

all: TapestryD N/TapestryNetworking

TapestryD: TapestryD.go
	$(CC) build TapestryD.go

N/TapestryNetworking: N/TapestryNetworking.go
	$(CC) build N/TapestryNetworking.go

