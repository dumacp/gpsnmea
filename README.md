# gpsnmea
read a nmea port serial

## Funcionamiento

El binario abre el puerto serie indicado y lee continuamente tramas NMEA. Filtra únicamente las sentencias `$GPRMC`, `$GNGNS` y `$GPVTG`. Si se activa la opción `-mqtt`, las tramas válidas se publican periódicamente (según el valor de `-timeout`) en el tópico `EVENTS/gps` del broker MQTT local en formato JSON.

Si el puerto serie se desconecta o falla, el programa espera 5 segundos y reintenta la conexión automáticamente.

---

## Instalación

### Requisitos previos

- [Go](https://golang.org/dl/) 1.16 o superior instalado.
- El directorio `$GOPATH/bin` (por defecto `$HOME/go/bin`) debe estar en la variable de entorno `PATH`.

Puedes verificarlo y agregarlo añadiendo esta línea a `~/.bashrc` o `~/.profile`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Instalar con `go install`

```bash
go install github.com/dumacp/gpsnmea/main@latest
```

Este comando descarga, compila e instala el binario automáticamente. El ejecutable quedará en:

```
$HOME/go/bin/main
```

### Ejecutar el binario

Si `$HOME/go/bin` está en el `PATH`, puedes invocarlo directamente:

```bash
main [opciones]
```

De lo contrario, usa la ruta completa:

```bash
~/go/bin/main [opciones]
```

---

## Uso

### Opciones disponibles

| Flag        | Valor por defecto | Descripción                                           |
|-------------|-------------------|-------------------------------------------------------|
| `-port`     | `/dev/ttyUSB1`    | Puerto serie desde el que se leen las tramas NMEA.    |
| `-baudRate` | `115200`          | Velocidad en baudios del puerto serie.                |
| `-timeout`  | `30`              | Segundos entre envíos de tramas al broker MQTT.       |
| `-mqtt`     | `false`           | Activa la publicación de mensajes al broker MQTT local.|

### Ejemplos

Leer el puerto `/dev/ttyUSB0` a 9600 baudios:

```bash
main -port /dev/ttyUSB0 -baudRate 9600
```

Leer el puerto y publicar en el broker MQTT local:

```bash
main -port /dev/ttyUSB0 -baudRate 115200 -mqtt
```

