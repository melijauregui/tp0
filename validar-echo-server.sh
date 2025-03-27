#!/bin/bash  
#bash como shell por defecto

SERVER_NAME="server"
NETWORK_NAME="tp0_testing_net"
MESSAGE="hello_world"

RESULT=$(docker run --rm --network="$NETWORK_NAME" alpine sh -c "echo '$MESSAGE' | nc $SERVER_NAME 12345")

#--network="$NETWORK_NAME": conecta el contenedor a la red de Docker creada en el compose
#alpine: imagen de Docker que contiene un sistema operativo mínimo visto en clase
#--rm: elimina el contenedor una vez que termina de ejecutar el comando
#sh abre una shell dentro del contenedor
#-c "comando" hace que sh ejecute el comando que está entre comillas
#echo '$MESSAGE' | nc $SERVER_NAME 12345: envía el mensaje a través de netcat al servidor

# Valida si el mensaje recibido es el mismo que se envió
if [ "$RESULT" == "$MESSAGE" ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi
