# TP0: Docker + Comunicaciones + Concurrencia

## Parte 1: Introducción a Docker

### Ejercicio N°1:

Se creó el `script generar-compose.sh`, el cual genera el archivo `docker-compose.yml` en la raíz del proyecto. Este script recibe como primer argumento el nombre del archivo de salida y como segundo argumento la cantidad de clientes a generar. A partir de estos parámetros, ejecuta el script `generate_docker_compose.py`, que se encarga de generar dinámicamente el contenido del archivo `docker-compose.yml` con la configuración correspondiente.


### Ejercicio N°2:
Consideré dos posibles soluciones para lograr que realizar cambios en el archivo de configuración no requiera reconstruír las imágenes de Docker para que los mismos sean efectivos. :

+ **Volume:** Los cambios en los archivos de configuración solo se ven reflejados si se copia manualmente el archivo local al volumen o si se editan directamente los archivos dentro del contenedor accediendo al volumen.

+ **Bind Mount:** Los cambios realizados en el archivo de configuración local se reflejan automáticamente dentro del contenedor al guardar los cambios en el host. Sin embargo, esta opción puede depender de la estructura del proyecto en el host y del sistema operativo utilizado.

Dado que se trata únicamente de dos archivos y la cátedra proporciona una estructura de proyecto predefinida, opté por utilizar la opción de bind mount para montar los archivos de configuración del cliente y del servidor dentro de los contenedores. Esta decisión fue discutida con los docentes en el [foro de la cátedra](https://campusgrado.fi.uba.ar/mod/forum/discuss.php?d=29503).