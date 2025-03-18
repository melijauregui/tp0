import socket

def send_message(sock, msg):
    """
    mensaje con formato `<longitud>:<msg>` y espera una respuesta.
    """
    msg_bytes = msg.encode()  
    total_length = len(msg_bytes)
    
    # asegura que se envía todo el mensaje
    sent = sock.send(msg_bytes)
    while sent < total_length:
        sent += sock.send(msg_bytes[sent:])

    return 

def read_message(sock):
    """
    Lee un mensaje con formato `<longitud>:<msg>` desde un socket.
    """
    # Leer hasta encontrar `:` (indica la longitud)
    length_str = b""
    while True:
        byte = sock.recv(1)  
        if byte == b":":
            break
        length_str += byte

    try:
        total_length = int(length_str.decode())  # convertir longitud a entero
    except ValueError:
        raise ValueError(f"Error al convertir la longitud: {length_str}")

    # leo exactamente `total_length` bytes
    received_bytes = b""
    while len(received_bytes) < total_length:
        chunk = sock.recv(total_length - len(received_bytes))
        if not chunk:
            raise ConnectionError("Conexión cerrada antes de recibir el mensaje completo")
        received_bytes += chunk

    return received_bytes.decode()  
