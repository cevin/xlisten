# XListen

Perceptionless systemd-socket inheritance implementation

Automatic inheritance when using systemd-socket

Otherwise, use the built-in `net` package to create the listener


# Example

```shell
# /etc/systemd/system/test.socket

[Socket]
ListenStream=0.0.0.0:8000

[Install]
WantedBy=sockets.target
```

```shell
# /etc/systemd/system/test.service

[Service]
Type=simple
ExecStart=/bin/any-executeable-binary-file

[Install]
WantedBy=multi-user.target
```

```go
package main

func main() {
	listener, err := xlisten.Listen("tcp4", ":8000")
}
```