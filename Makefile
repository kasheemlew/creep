bench:
	benchmark -c 10000 -n 1000000 -proxy socks5://127.0.0.1:1080 http://127.0.0.1:8080/ping
