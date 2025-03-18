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
	//enviar mensaje
	n_write, err := fmt.Fprint(conn, msg)
	if err != nil {
		return "", err
	}

	//asegurarse de que se envió todo el mensaje
	for n_write < len(msg) {
		n_write2, err := fmt.Fprint(conn, msg[n_write:])
		if err != nil {
			return "", err
		}
		n_write += n_write2
	}

	//leer respuesta
	return ReadMessage(conn, msg)
}

// mensaje con formato `<longitud>:<msg>`
func ReadMessage(conn net.Conn, msg string) (string, error) {
	reader := bufio.NewReader(conn)
	lengthStr, err := reader.ReadString(':')
	if err != nil {
		return "", err
	}

	//elimino el `:` del final
	lengthStr = strings.TrimSuffix(lengthStr, ":")
	totalLength, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", fmt.Errorf("error al convertir la longitud: %v", err)
	}

	log.Infof("Longitud del mensaje: %d\n", totalLength)

	//Leer exactamente `totalLength` bytes
	bs := make([]byte, totalLength)
	bytesRead := 0
	for bytesRead < totalLength {
		n, err := reader.Read(bs[bytesRead:])
		if err != nil {
			return "", fmt.Errorf("error al leer mensaje completo: %v", err)
		}
		log.Infof("Longitud leida: %d\n", n)
		bytesRead += n
	}

	receivedMsg := string(bs[:bytesRead])

	return receivedMsg, nil
}
