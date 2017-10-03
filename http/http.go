package http

import "bytes"

func Accept(key string) []byte {
	var buffer bytes.Buffer
	buffer.WriteString("HTTP/1.1 101 Switching Protocols\r\n")
	buffer.WriteString("Upgrade: websocket\r\n")
	buffer.WriteString("Connection: Upgrade\r\n")
	buffer.WriteString("Sec-WebSocket-Accept: ")
	buffer.WriteString(key)
	buffer.WriteString("\r\n\r\n")
	return buffer.Bytes()
}
