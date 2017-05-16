CC=go

all: TapestryD N/TapestryNetworking

TapestryD: tapestry_d.go
	$(CC) build -o TapestryD tapestry_d.go

N/TapestryNetworking: N/tapestry_networking.go
	$(CC) build -o TapestryNetworking N/tapestry_networking.go

