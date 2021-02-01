test:
	go test ./pkg/...

fmt:
	go fmt ./cmd/... ./pkg/...

vet:
	go vet ./cmd/... ./pkg/...

cyclonus:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./cmd/cyclonus/cyclonus ./cmd/cyclonus
	docker build -t docker.io/mfenwick100/cyclonus:latest ./cmd/cyclonus
