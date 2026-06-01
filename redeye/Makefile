all: 
	go build -v

pi:
	env GOOS=linux GOARCH=arm GOARM=7 go build -v

clean:
	rm -f $(CMD) *.log
	go clean
