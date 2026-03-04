.PHONY: build frontend backend dev clean

build: frontend backend

frontend:
	cd frontend && npm install && npm run build

backend:
	CGO_ENABLED=1 go build -o ./bin/flash-cards ./cmd/server

dev:
	./dev.sh

clean:
	rm -rf bin/ frontend/dist/ frontend/node_modules/
