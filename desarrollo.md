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
En el servidor, se registró el método `graceful_shutdown` como handler para las señales `SIGTERM` y `SIGINT` mediante el módulo signal de Python. Este método, se encarga del cierre tanto la conexión actual con el cliente (`self._client_socket`) como el socket (`self._server_socket`) que utiliza para escuchar nuevas conexiones. Además, setea el valor del atributo `self._running` en False lo que permite que el hilo principal finalice su ejecución de forma controlada (graceful).

En el caso del cliente, el handler de señales se ejecuta en una goroutine independiente y se encarga de escuchar la señal `SIGTERM` del sistema operativo. Para ello, utiliza un canal `sigChannel` junto con `signal.Notify`, lo que permite detectar si el proceso fue terminado de forma externa, ejecutando los métodos correspondientes para exitear cada cliente y servidor de forma graceful. 
Al recibirla, el cliente actualiza su atributo `running` a `false` (igual que el servidor) el cual detiene el envío de mensajes y, si existe una conexión activa con el servidor (`c.conn`), la cierra.
Además, se incorporó un canal `finishChan` que permite finalizar la goroutine encargada de manejar señales en caso de que el proceso finalice de forma normal, sin recibir una señal `SIGTERM`. De esta manera, si no se recibe dicha señal, se envía una notificación a través de `finishChan` que indica que la ejecución ha concluido, permitiendo cerrar correctamente la goroutine del handler y esperar su finalización mediante un `WaitGroup`. Esto asegura una terminación ordenada del programa y evita dejar goroutines colgadas.

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

### Ejercicio N°6:

Para este ejercicio, cada mensaje enviado al servidor corresponde a un **batch message**, cuyo formato es similar al utilizado en el ejercicio 5, con la diferencia de que ahora se agrupan múltiples apuestas en un mismo mensaje, separadas por punto y coma (`;`). El formato general es:

```
apuesta1;apuesta2;...;apuesta_n
```

donde `n` representa la cantidad máxima de apuestas por batch, o una cantidad menor si no quedan más apuestas por enviar.

El cliente procesa el archivo `client_{id}.csv` leyendo línea por línea y concatenando las apuestas en un único string que representa el mensaje del batch a enviar al servidor.

Para definir el valor por defecto de la cantidad de apuestas por batch (`batch.maxAmount`), se asumió que cada línea del CSV ocupa como máximo 80 caracteres. Esta estimación se basa en los ejemplos provistos, donde incluso las líneas más largas no superan dicha longitud. Dado que el tamaño máximo de un mensaje es de 8 KB, se estableció un valor por defecto de 100 apuestas por batch, ya que este límite garantiza que el mensaje no supere el tamaño permitido.

Además, se incorporó un nuevo atributo `fileReader` al cliente para mantener una referencia al archivo abierto, lo que permite cerrarlo adecuadamente al momento de realizar un shutdown de manera *graceful*.

Por cada batch enviado, el servidor responde con un mensaje `{cantidad} apuestas almacenadas` indicando la cantidad de apuestas que fueron almacenadas correctamente. Esta respuesta permite al cliente verificar si todas las apuestas del batch fueron recibidas con éxito. En función de esta validación, el cliente imprime:

+ `action: apuesta_enviada | result: success`, si la cantidad almacenada coincide con la cantidad enviada.

+ `action: apuesta_enviada | result: fail`, si hubo algún error en el almacenamiento de alguna apuesta.


### Ejercicio N°7:
Para este ejercicio, cada cliente envía su batch de apuestas de la misma manera que en el ejercicio 6. La diferencia principal es que, una vez finalizado el envío de todas sus apuestas, el cliente envía al servidor el mensaje `"Winners, please?"`, indicando que ha concluido su participación.

El servidor puede responder de dos maneras:

+ `"No winners yet"`: si aún el servidor no ha determinado a los ganadores, es decir, si todavía hay agencias que no han finalizado el envío de sus apuestas.
+ `"ganador1;ganador2;...;ganador_n"`: una cadena que representa la lista de documentos de las apuestas ganadoras correspondientes a esa agencia.

Del lado del servidor, se mantiene un map (`agenciesWaiting`) que registra las agencias que han finalizado el envío de apuestas. Cuando el servidor recibe el mensaje `"Winners, please?"` por parte de una agencia, la agrega al map (indicando que ya finalizo su envio de apuestas). Una vez que todas las agencias han sido registradas como finalizadas, el servidor procede a procesar los ganadores del sorteo. El resultado se almacena en el mismo map, donde la clave es el ID de la agencia y el valor es el mensaje con los documentos ganadores correspondiente a esa agencia (`"ganador1;ganador2;...;ganador_n"`).

Por último, si el cliente recibe como respuesta `"No winners yet`", se le penaliza con un segundo adicional de espera antes de volver a consultar, incrementando progresivamente el tiempo entre cada solicitud de resultados hasta que estos estén disponibles.

## Parte 3: Repaso de Concurrencia
### Ejercicio N°8:

Para implementar la concurrencia en el servidor, se modificó la ejecución del método `handleClientConnection` para que se ejecute dentro de una goroutine, permitiendo atender múltiples clientes en paralelo mientras el servidor continúa aceptando nuevas conexiones.

Esto requirió realizar una serie de cambios en la estructura Server para asegurar el manejo correcto de múltiples conexiones y garantizar una finalización ordenada en caso de shutdown, todo de manera thread-safe. Las principales modificaciones fueron:

+ Se reemplazó el campo `s.clientCon`, que almacenaba una única conexión, por un map` s.clientsConn` que permite registrar múltiples conexiones activas. Para garantizar acceso seguro desde múltiples goroutines, se incorporó un mutex asociado a este map (`s.lockClientsConn`).
+ Se protegió el acceso al atributo `s.running` mediante un mutex, permitiendo leer y modificar su valor de manera thread-safe desde diferentes partes del código.
+ Se adaptó el método `GracefulShutdown()` del servidor para cerrar todas las conexiones activas almacenadas en `s.clientsConn` y establecer `s.running = false`, utilizando los mutex correspondientes para evitar condiciones de carrera.
+ Se introdujo el mutex `s.betsLock` para proteger las operaciones de lectura y escritura sobre las estructuras de datos que almacenan las apuestas, ya que estas operaciones no son seguras por defecto en contextos concurrentes.
+ Se creó el mutex `s.lockWinnerRevealed` para sincronizar el acceso a la estructura `agenciesWaiting`, que guarda el conjunto de agencias que finalizaron el envío de apuestas. Este lock se utiliza tanto al registrar nuevas agencias (cuando el servidor recibe un mensaje `"Winners, please?"`) como al consultar cuántas agencias están en espera, evitando así interferencias entre el hilo principal y las goroutines que manejan las conexiones.
+ Finalmente, cada goroutine creada a través del `handleClientConnection` se registra en un `WaitGroup`, que es utilizado por el hilo principal para esperar a que todas las conexiones activas finalicen antes de completar el shutdown, garantizando así una finalización ordenada del servidor.

## Diagramas protocolo de comunicación

![alt text](image.png)
![alt text](image-1.png)
![alt text](image-2.png)
![alt text](image-3.png)