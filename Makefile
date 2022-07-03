build:
	go build -o bin/cmd cmd/main.go
run: build
	./bin/main
watch:
	reflex -s -r '\.go$$' make run