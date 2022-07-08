build:
	go build -o bin/cmd cmd/main.go
run: build
	./bin/cmd
watch:
	reflex -s -r '\.go$$' make run