# GoRelay
GPIO controller via HTTP with golang and periph.io

# Usage
```bash
go get
go run .
```
```bash
# https://www.raspberrypi.com/documentation/computers/os.html#gpio-and-the-40-pin-header
curl -X PUT "http://127.0.0.1:8080/pin/17/1"
curl -X PUT "http://127.0.0.1:8080/pin/27/0"
# https://github.com/bigtreetech/CB1#40-pin-gpio
curl -X PUT "http://127.0.0.1:8080/pin/78/1"
curl -X PUT "http://127.0.0.1:8080/pin/78/0"
```
