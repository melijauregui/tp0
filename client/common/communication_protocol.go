package common

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// mensaje con formato `<longitud>:<msg>` y recibe respuesta
func SendMessage(conn net.Conn, msg string) (string, error) {
	msgSend := fmt.Sprintf("%d:%s", len(msg), msg)
	//enviar mensaje
	n_write, err := fmt.Fprint(conn, msgSend)
	if err != nil {
		return "", err
	}

	//asegurarse de que se envió todo el mensaje
	for n_write < len(msgSend) {
		n_write2, err := fmt.Fprint(conn, msgSend[n_write:])
		if err != nil {
			return "", err
		}
		n_write += n_write2
	}

	//leer respuesta
	return ReadMessage(conn, msgSend)
}

// mensaje con formato `<longitud>:<msg>`
func ReadMessage(conn net.Conn, msg string) (string, error) {
	reader := bufio.NewReader(conn)
	totalBytesStr, err := reader.ReadString(':')
	if err != nil {
		return "", err
	}

	//elimino el `:` del final
	totalBytesStr = strings.TrimSuffix(totalBytesStr, ":")
	totalBytes, err := strconv.Atoi(totalBytesStr)
	if err != nil {
		return "", fmt.Errorf("error al convertir la longitud: %v", err)
	}

	//Leer exactamente `totalLength` bytes
	bs := make([]byte, totalBytes)
	bytesRead := 0
	for bytesRead < totalBytes {
		n, err := reader.Read(bs[bytesRead:])
		if err != nil {
			return "", fmt.Errorf("error al leer mensaje completo: %v", err)
		}
		bytesRead += n
	}

	receivedMsg := string(bs[:bytesRead])

	return receivedMsg, nil
}
