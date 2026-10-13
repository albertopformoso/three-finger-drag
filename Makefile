.PHONY: fmt test probe

fmt:
	pkg-config --atleast-version=1.14.0 libevdev
	gofmt -w experiments/inputprobe/*.go
	go vet ./...

test:
	go test ./...

probe:
	pkg-config --atleast-version=1.14.0 libevdev
	go build -o /tmp/opencode/three-finger-inputprobe ./experiments/inputprobe
