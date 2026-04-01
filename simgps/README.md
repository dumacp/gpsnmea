# simgps

Simulador de GPS que genera tramas `$GPRMC` (NMEA) a partir de una lista de puntos geográficos en formato JSON. Útil para pruebas sin hardware GPS real.

## Funcionamiento

`simgps` lee puntos GPS (latitud/longitud) desde un archivo JSON o desde la entrada estándar (stdin), y por cada punto genera una trama `$GPRMC` válida con checksum, incluyendo fecha y hora actuales del sistema.

Las tramas se imprimen en la salida estándar (`stdout`) y, opcionalmente, se publican en un broker MQTT en los tópicos `EVENTS/GPS` y `GPS`.

- **Modo archivo**: lee un arreglo JSON de puntos completo del archivo indicado y emite todas las tramas espaciadas por el valor de `-timeout`.
- **Modo stdin**: lee un punto JSON por línea/entrada y emite la trama correspondiente en tiempo real. El ritmo de emisión lo controla el proveedor de datos desde stdin.

---

## Instalación

### Requisitos previos

- [Go](https://golang.org/dl/) 1.16 o superior instalado.
- El directorio `$GOPATH/bin` (por defecto `$HOME/go/bin`) debe estar en el `PATH`.

Agrega esta línea a `~/.bashrc` o `~/.profile` si aún no lo tienes:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Instalar con `go install`

```bash
go install github.com/dumacp/gpsnmea/simgps@latest
```

El binario quedará en:

```
$HOME/go/bin/simgps
```

### Ejecutar el binario

```bash
simgps [opciones] [FILE]
```

O con ruta completa si el `PATH` no está configurado:

```bash
~/go/bin/simgps [opciones] [FILE]
```

### Binarios precompilados

En el directorio [`outputs/`](outputs/) se encuentran binarios precompilados listos para usar en determinadas arquitecturas:

| Archivo             | Arquitectura         |
|---------------------|----------------------|
| `simgps_armv7`      | Linux ARM v7 (32-bit)|

Para usarlos directamente, descarga el binario correspondiente, dale permisos de ejecución y ejecútalo:

```bash
chmod +x outputs/simgps_armv7
./outputs/simgps_armv7 [opciones] [FILE]
```

### Compilación cruzada (cross-compile)

Si necesitas compilar para una arquitectura diferente, puedes usar las variables de entorno de Go para realizar cross-compilation. Por ejemplo, para generar el binario ARM v7 incluido en `outputs/`:

```bash
env GOOS=linux GOARCH=arm GOARM=7 go build -o outputs/simgps_armv7
```

Otros ejemplos de arquitecturas comunes:

| Arquitectura         | Comando                                                              |
|----------------------|----------------------------------------------------------------------|
| Linux ARM v7 (32-bit)| `env GOOS=linux GOARCH=arm GOARM=7 go build -o outputs/simgps_armv7`|
| Linux ARM64          | `env GOOS=linux GOARCH=arm64 go build -o outputs/simgps_arm64`      |
| Linux x86-64         | `env GOOS=linux GOARCH=amd64 go build -o outputs/simgps_amd64`      |
| Windows x86-64       | `env GOOS=windows GOARCH=amd64 go build -o outputs/simgps.exe`      |

Puedes consultar todas las combinaciones soportadas con:

```bash
go tool dist list
```

---

## Ayuda de la CLI

Puedes consultar el uso directamente desde el binario:

```bash
simgps -h
```

Salida:

```
Uso de simgps:  <binary> [OPTION...] [FILE]

Genera tramas GPRMC a partir de un archivo (FILE) JSON (itinerario desde plataforma) con puntos GPS

Si FILE no se especifica, se leerá desde la entrada estándar, y los datos de entradas deben ser
puntos, no un arreglo de puntos como en el caso del archivo (FILE). Cuando se usa la entrada
estándar (STDIN) el timeout será controlado por el ingreso de datos desde la STDIN.

Opciones:
  -ip string
        ip address for mqtt broker (default "127.0.0.1")
  -mqtt
        enable mqtt send to broker
  -port int
        port for mqtt broker (default 1883)
  -timeout duration
        timeout for the request (default 1s)
```

---

## Formato del archivo JSON de entrada

El archivo JSON de entrada es un **itinerario exportado desde la plataforma**, que contiene un arreglo de puntos de control (checkpoints) con sus coordenadas GPS y metadatos asociados:

```json
[
  {
    "checkPointId": "1",
    "type": "stop",
    "name": "Terminal Norte",
    "radios": 50,
    "maxSpeed": "60",
    "lat": 6.3,
    "long": -75.9
  },
  {
    "checkPointId": "2",
    "type": "waypoint",
    "name": "Cruce Central",
    "radios": 30,
    "maxSpeed": "40",
    "lat": 6.98,
    "long": -75.87
  }
]
```

Los campos mínimos requeridos son `lat` y `long`. Los demás campos (`checkPointId`, `type`, `name`, `radios`, `maxSpeed`) son opcionales para la generación de tramas.

Cuando se usa **stdin**, se envía un objeto JSON por vez (sin arreglo), por ejemplo al consumir una fuente en tiempo real:

```json
{ "lat": 6.3, "long": -75.9 }
```

---

## Opciones

| Flag      | Valor por defecto | Descripción                                                    |
|-----------|-------------------|----------------------------------------------------------------|
| `-timeout`| `1s`              | Tiempo de espera entre tramas al leer desde archivo.           |
| `-mqtt`   | `false`           | Activa la publicación de tramas en el broker MQTT.             |
| `-ip`     | `127.0.0.1`       | Dirección IP del broker MQTT.                                  |
| `-port`   | `1883`            | Puerto del broker MQTT.                                        |

---

## Ejemplos

### Leer desde un archivo JSON

```bash
simgps -timeout 500ms itinerario.json
```

### Leer desde stdin (punto a punto)

```bash
echo '{"lat": 6.3, "long": -75.9}' | simgps
```

### Generar tramas y publicarlas en MQTT

```bash
simgps -mqtt -ip 192.168.1.10 itinerario.json
```

### Encadenar con otro proceso usando pipes

```bash
cat itinerario.json | jq -c '.[]' | simgps -mqtt
```

### Loop infinito enviando el mismo punto cada 2 segundos

```bash
while true; do
  echo '{"lat": 6.3, "long": -75.9}'
  sleep 2
done | simgps
```

Con publicación MQTT:

```bash
while true; do
  echo '{"lat": 6.3, "long": -75.9}'
  sleep 2
done | simgps -mqtt -ip 192.168.1.10
```

---

## Tópicos MQTT publicados

Cuando `-mqtt` está activo, por cada trama se publican dos mensajes:

| Tópico       | Formato del mensaje                                                                 |
|--------------|-------------------------------------------------------------------------------------|
| `EVENTS/GPS` | `{"timeStamp": <unix_seconds_float>, "value": "<trama_GPRMC>", "type": "GPRMC"}`  |
| `GPS`        | Trama `$GPRMC` en texto plano.                                                      |
