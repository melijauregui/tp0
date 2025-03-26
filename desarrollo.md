# TP0: Docker + Comunicaciones + Concurrencia

## Parte 1: Introducción a Docker

### Ejercicio N°1:

Se creó el `script generar-compose.sh`, el cual genera el archivo `docker-compose.yml` en la raíz del proyecto. Este script recibe como primer argumento el nombre del archivo de salida y como segundo argumento la cantidad de clientes a generar. A partir de estos parámetros, ejecuta el script `generate_docker_compose.py`, que se encarga de generar dinámicamente el contenido del archivo `docker-compose.yml` con la configuración correspondiente.

### Ejercicio N°2:
Consideré dos posibles soluciones para lograr que realizar cambios en el archivo de configuración no requiera reconstruír las imágenes de Docker para que los mismos sean efectivos. :

+ **Volume:** Los cambios en los archivos de configuración solo se ven reflejados si se copia manualmente el archivo local al volumen o si se editan directamente los archivos dentro del contenedor accediendo al volumen.

+ **Bind Mount:** Los cambios realizados en el archivo de configuración local se reflejan automáticamente dentro del contenedor al guardar los cambios en el host. Sin embargo, esta opción puede depender de la estructura del proyecto en el host y del sistema operativo utilizado.

Dado que se trata únicamente de dos archivos y la cátedra proporciona una estructura de proyecto predefinida, opté por utilizar la opción de bind mount para montar los archivos de configuración del cliente y del servidor dentro de los contenedores. Esta decisión fue discutida con los docentes en el [foro de la cátedra](https://campusgrado.fi.uba.ar/mod/forum/discuss.php?d=29503).

### Ejercicio N°3:
Para el script `validar-echo-server.sh` se utilizó el comando docker run sobre una imagen liviana de `Alpine` (imagen de Docker que contiene un sistema operativo mínimo visto en clase) que tiene `netcat` instalado. El objetivo fue verificar que el servidor estuviera funcionando correctamente, enviando el mensaje **"hello_world"** al puerto 12345 del servidor y validando que la respuesta recibida fuera exactamente la misma.

El contenedor se conecta a la red de Docker definida en el `docker-compose` (`tp0_testing_net`), y se lo elimina automáticamente una vez ejecutado el comando. El test se considera exitoso si la respuesta coincide con el mensaje enviado.

En caso de validación exitosa, el script imprime:

```
action: test_echo_server | result: success
```

De lo contrario, imprime:

```
action: test_echo_server | result: fail
```


### Ejercicio N°4:
Se modificó el servidor para que, al recibir la señal `SIGTERM`, cierre tanto la conexión actual con el cliente (`self._client_socket`) como el socket (`self._server_socket`) que utiliza para escuchar nuevas conexiones. Además, el handler de esta señal, setea el valor del atributo `self._running` en False lo que permite que el hilo principal finalice su ejecución de forma controlada (graceful)

En el caso del cliente, se implementó una goroutine que escucha la señal `SIGTERM`. Al recibirla, el cliente actualiza su atributo `running` a `false` (igual que el servidor) el cual detiene el envío de mensajes y, si existe una conexión activa con el servidor (`c.conn`), la cierra. Esto permite que el cliente finalice su ejecución de forma graceful, liberando correctamente los recursos utilizados.

## Parte 2: Repaso de Comunicaciones

### Ejercicio N°5:
#### Protocolo de comunicación

El protocolo de comunicación utilizado para el envío y recepción de paquetes se encuentra implementado en los archivos `communication_protocol.go` (cliente) y `communication_protocol.py` (servidor).  
Este protocolo define que los mensajes se envían como cadenas de texto con el siguiente formato:

```
<longitud_del_mensaje:mensaje>
```

La inclusión del prefijo de longitud permite al receptor saber exactamente cuántos bytes debe leer, lo cual facilita una correcta delimitación de los mensajes.
Además, este diseño simplifica la implementación de la lógica necesaria para manejar short reads y short writes, que pueden ocurrir cuando las operaciones de lectura o escritura no transfieren todos los bytes esperados en una única llamada.

#### Serialización del mensaje de apuesta

El mensaje de apuesta se serializa como una cadena de texto donde los campos están separados por comas. El formato utilizado es el siguiente:

```
agencia,nombre,apellido,dni,nacimiento,numero
```

#### Serialización de la respuesta del servidor

La respuesta del servidor ante la recepción de una apuesta también se envía como una cadena de texto. Puede tomar uno de los siguientes valores:

- `"Apuesta almacenada"`: en caso de que el mensaje haya sido recibido y procesado correctamente.
- `"Apuesta no almacenada"`: si ocurrió algún error durante el envío o procesamiento del mensaje.

Esta respuesta permite al cliente validar si la apuesta fue registrada con éxito o si es necesario reintentar o informar un error.
