ECHO = @echo
GO = go

help:
	$(ECHO) "Raspberry Pi Beerpong back Makefile"
	$(ECHO) "'make go-build'		Build the API, don't forget to source your custom configuration in environment variables"
	$(ECHO) "'make go-build-arm'	Build for Linux ARMV7"

go-build:
	$(GO) build -ldflags "-s -w -X main.Version=$(VERSION)"

go-build-arm:
	GOOS=linux GOARCH=arm GOARM=7 $(GO) build -ldflags "-s -w -X main.Version=$(VERSION)"
