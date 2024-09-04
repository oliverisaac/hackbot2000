build: 
	env GOOS=linux GOARCH=amd64 go build .

format:
	gofmt -w *.go

copy: build format
	scp ./hackbot2000 isaacinit.com:/app/bin/hackbot2000.new
	ssh isaacinit.com systemctl stop hackbot2000
	ssh isaacinit.com mv /app/bin/hackbot2000.new /app/bin/hackbot2000
	ssh isaacinit.com systemctl start hackbot2000
	ssh isaacinit.com systemctl status hackbot2000

run: copy
