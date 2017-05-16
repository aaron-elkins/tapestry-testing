CC=go

all: TapestryD N/TapestryNetworking

TapestryD: tapestryd.go
	$(CC) build -o TapestryD tapestryd.go

N/TapestryNetworking: N/tapestry_networking.go
	$(CC) build -o TapestryNetworking N/tapestry_networking.go

