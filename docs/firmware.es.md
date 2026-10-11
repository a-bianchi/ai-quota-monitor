# Firmware y pantalla

[English](firmware.md) · **Español**

Firmware para la **Waveshare ESP32-C6-LCD-1.47** (sin touch). Se conecta a tu Wi-Fi, lee `GET /usage` de tu propio servidor y muestra los cupos personales y laborales, cambiando de página cada 10 segundos.

Hay dos sketches, uno por orientación de la pantalla: [`firmware/esp32-c6-lcd-1.47/ai_quota_monitor_vertical/`](../firmware/esp32-c6-lcd-1.47/ai_quota_monitor_vertical/) (172 × 320) y [`firmware/esp32-c6-lcd-1.47/ai_quota_monitor_horizontal/`](../firmware/esp32-c6-lcd-1.47/ai_quota_monitor_horizontal/) (320 × 172). Los dos muestran una cuenta por página, con letra grande. Subí a la placa el que corresponda a cómo la vas a apoyar.

![Pantallas renderizadas en una computadora con el código de dibujo del firmware](assets/firmware-preview.png)

La imagen es un render hecho en una computadora con el código de dibujo del firmware y los datos mock de la API. No es una foto de la placa.

## Estado de verificación

| Qué | Estado |
|---|---|
| Demo oficial de Waveshare en la placa real (compilar, subir, pantalla encendida) | Probado |
| Este firmware compila para ESP32C6 Dev Module con las versiones de abajo | Verificado (unos 575 KB) |
| Parser y diseño de pantalla contra respuestas reales de `GET /usage`, en una computadora | Verificado |
| Este firmware en la placa física: pantalla, Wi-Fi, HTTP | Probado |
| Este firmware en la placa física con una dirección `https://` | **Todavía no probado** |

## Versiones

Usá exactamente estas versiones; no las mezcles.

| Componente | Versión | De dónde sale |
|---|---|---|
| Arduino IDE | 2.3.10 | <https://www.arduino.cc/en/software> |
| esp32 by Espressif Systems | 3.3.12 | Gestor de placas de Arduino IDE |
| LVGL | 8.3.10 | `Arduino/libraries/lvgl` del ZIP del demo de Waveshare |
| ArduinoJson | 7.4.3 | Gestor de bibliotecas de Arduino IDE |

LVGL 9 no funciona con el demo de esta placa ni con este firmware. PNGdec 1.0.2 también viene en el ZIP; el demo oficial puede necesitarla, este firmware no.

Fuentes oficiales:

- Documentación de la placa: <https://www.waveshare.com/wiki/ESP32-C6-LCD-1.47>
- Demo oficial: <https://files.waveshare.com/wiki/ESP32-C6-LCD-1.47/ESP32-C6-LCD-1.47-Demo.zip>
- Índice de placas ESP32 para Arduino: <https://espressif.github.io/arduino-esp32/package_esp32_index.json>

## Paso a paso

No hace falta experiencia previa con Arduino. Seguí los pasos en orden: del 1 al 11 se enciende la pantalla con el demo del fabricante, lo que comprueba que la placa, el cable y las herramientas funcionan antes de sumar Wi-Fi y la API.

### 1. Hardware y cable

- Una placa Waveshare ESP32-C6-LCD-1.47 (no la S3 ni la versión Touch). Es la placa sobre la que se desarrolló el proyecto; ver [dónde comprarla](hardware.es.md#placa-usada-y-dónde-comprarla).
- Un cable USB-C **de datos**. Muchos cables solo cargan: con esos la placa enciende, pero la computadora nunca la detecta.

La pantalla viene integrada en la placa. No hay nada que cablear y no se necesita microSD.

### 2. Descargas

Descargá estas dos cosas:

- Arduino IDE, desde <https://www.arduino.cc/en/software>.
- El ZIP del demo oficial, desde <https://files.waveshare.com/wiki/ESP32-C6-LCD-1.47/ESP32-C6-LCD-1.47-Demo.zip> (unos 60 MB).

### 3. Instalar Arduino IDE

Instalalo como cualquier otro programa y abrilo una vez.

### 4. Instalar las placas ESP32

1. Abrí **Herramientas → Placa → Gestor de placas**.
2. Buscá `esp32`.
3. En **esp32 by Espressif Systems**, elegí la versión **3.3.12** y presioná **Instalar**. Tarda varios minutos.

Si `esp32` no aparece, abrí las **Preferencias** de Arduino IDE, pegá esta dirección en **URLs adicionales de gestor de placas** y probá de nuevo:

```text
https://espressif.github.io/arduino-esp32/package_esp32_index.json
```

### 5. Extraer el demo

Extraé el ZIP en cualquier lugar. Las rutas de abajo son relativas a su contenido:

- `Arduino/examples/LVGL_Arduino/`: el ejemplo de referencia.
- `Arduino/libraries/`: las bibliotecas, `lvgl` y `PNGdec`.

### 6. Instalar las bibliotecas

La biblioteca LVGL tiene que ser la del ZIP: incluye `src/lv_conf.h`, ya configurado para esta placa.

1. Abrí las **Preferencias** de Arduino IDE.
2. Consultá la **Ruta del Sketchbook**.
3. Abrí esa carpeta y creá adentro una carpeta `libraries` si no existe.
4. Copiá las carpetas completas `lvgl` y `PNGdec` de `Arduino/libraries/` del ZIP a esa carpeta `libraries`.
5. Cerrá y volvé a abrir Arduino IDE.

Si más adelante Arduino IDE ofrece actualizar LVGL, no aceptes: la versión 9 no funciona acá.

Después instalá el parser JSON, que no viene en el ZIP:

1. Abrí **Herramientas → Administrar bibliotecas**.
2. Buscá `ArduinoJson` (de Benoit Blanchon).
3. Elegí la versión **7.4.3** y presioná **Instalar**.

### 7. Abrir el demo oficial

Abrí **Archivo → Abrir** y elegí `Arduino/examples/LVGL_Arduino/LVGL_Arduino.ino` del ZIP extraído.

### 8. Elegir la placa y el puerto

1. Conectá la placa con el cable USB-C.
2. **Herramientas → Placa → esp32 → ESP32C6 Dev Module**.
3. **Herramientas → Puerto**: elegí el puerto que aparece al enchufar la placa. El nombre depende de tu sistema; desenchufá y enchufá la placa para ver cuál es.

### 9. Elegir el esquema de particiones

**Herramientas → Partition Scheme → Huge APP (3MB No OTA/1MB SPIFFS)**.

Sin eso, el demo no entra en la memoria de la placa.

### 10. Verificar y subir

Los dos botones de arriba a la izquierda hacen cosas distintas:

- **✓ Verificar** solo compila. Comprueba que el programa y las bibliotecas estén bien, y no toca la placa.
- **→ Subir** compila y después graba el programa en la placa.

Presioná primero **✓ Verificar**. La primera vez tarda varios minutos. Cuando termine sin errores, presioná **→ Subir**.

### 11. Confirmar la carga

La carga terminó bien cuando el panel de salida, abajo, muestra estas dos líneas:

```text
Hash of data verified.
Hard resetting via RTS pin...
```

La placa se reinicia sola y la pantalla muestra el demo del fabricante. Si lo hace, la placa, el cable y las herramientas funcionan.

### 12. Abrir el sketch de este proyecto

Conseguí el repositorio (con `git clone`, o **Code → Download ZIP** en la página del repositorio) y, en Arduino IDE, abrí uno de los dos sketches:

```text
firmware/esp32-c6-lcd-1.47/ai_quota_monitor_vertical/ai_quota_monitor_vertical.ino
firmware/esp32-c6-lcd-1.47/ai_quota_monitor_horizontal/ai_quota_monitor_horizontal.ino
```

Lo único que cambia entre ellos es `layout.h`. Ahí también podés girar la imagen 180° (`DISPLAY_FLIP 1`) si la placa queda al revés, o volver a la vista con todas las cuentas de un grupo en letra chica (`UI_ONE_ACCOUNT 0`).

La carpeta y el archivo `.ino` tienen que conservar el mismo nombre. No saques el `.ino` de su carpeta: los archivos que están al lado son parte del programa.

### 13. Configurar Wi-Fi, API y token

1. En la carpeta del sketch que abriste, copiá `config.example.h` a un archivo nuevo llamado `config.h`, en esa misma carpeta.
2. Abrí `config.h` y completá los valores. Arduino IDE lo muestra como una pestaña cuando existe; si no aparece, volvé a abrir el sketch.

| Ajuste | Qué poner |
|---|---|
| `WIFI_SSID` | Nombre de tu red Wi-Fi. Tiene que ser de 2,4 GHz |
| `WIFI_PASSWORD` | Su contraseña |
| `API_BASE_URL` | Dirección de **tu** servidor: `http://IP_O_HOSTNAME:PUERTO`, sin barra final |
| `DEVICE_TOKEN` | El `DEVICE_TOKEN` de ese servidor |
| `UI_LANGUAGE` | `"es"` o `"en"` |
| `API_TIMEZONE` | Zona horaria IANA para las horas en pantalla, por ejemplo `Europe/Madrid` |
| `API_CA_CERT` | Solo para una autoridad certificante privada con `https://`; ver [HTTPS](#https) |
| `DISPLAY_BRIGHTNESS` | 0 a 100. El fabricante recomienda 50 o menos |

`API_BASE_URL` apunta a donde hayas desplegado el servidor: la IP y el puerto son los de tu propia instalación, no un valor fijo. Dos ejemplos:

- Servidor levantado con `make docker-up` en una computadora con IP `192.168.1.50` y `HOST_BIND=0.0.0.0`: `http://192.168.1.50:8080`.
- Servidor desplegado con `compose.coolify.yaml`, que publica el puerto 8090: `http://IP_DEL_SERVIDOR:8090`.

**`localhost` no sirve.** En la ESP32, `localhost` y `127.0.0.1` son la propia ESP32, no tu computadora. Usá la IP LAN o el hostname del equipo que corre el servidor.

`config.h` está ignorado por Git. En la placa va solo el token de lectura: nunca pongas ahí keys de providers, sesiones de Codex o Claude, `COLLECTOR_TOKEN` ni la contraseña del panel.

### 14. Ejecutar la API y acceder desde la LAN

El servidor tiene que estar corriendo y aceptar conexiones desde la red, no solo desde sí mismo. Ver [despliegue](deployment.es.md): con Docker Compose eso significa poner `HOST_BIND` en la IP LAN del servidor o en `0.0.0.0`.

Antes de subir, comprobá desde **otro** equipo de la misma red que la dirección responde. Reemplazá la dirección y el token por los tuyos:

```sh
curl http://IP_O_HOSTNAME:PUERTO/ping

curl -H 'Authorization: Bearer TU_DEVICE_TOKEN' \
  http://IP_O_HOSTNAME:PUERTO/usage
```

El primero tiene que responder `{"status":"ok"}` y el segundo un JSON con `accounts`. Si no funcionan desde otra computadora, tampoco van a funcionar desde la placa.

### 15. Subir y resultado esperado

Con la misma placa, puerto y esquema de particiones de los pasos 8 y 9, presioná **✓ Verificar** y después **→ Subir**. Esperá `Hash of data verified.` y `Hard resetting via RTS pin...`.

Lo que muestra la pantalla:

1. Durante unos segundos, **Sin Wi-Fi** mientras se conecta.
2. Después la página **PERSONAL** en celeste y, 10 segundos más tarde, la página **TRABAJO** en violeta. Siguen alternando. Si las cuentas de un grupo no entran, ese grupo usa más páginas y arriba a la derecha se indica cuál es, por ejemplo `2/3`. Un grupo sin cuentas se saltea.
3. Por cada cuenta: su nombre, cada cupo con lo que queda, una barra con el uso, la renovación cuando existe y la última actualización de esa cuenta.
4. Abajo, cuándo se leyó la API por última vez.

## Qué muestra la pantalla

- **Colores**: según el `usageLevel` de cada cupo: `low` verde, `medium` naranja, `high` rojo, `unknown` gris. La placa no calcula porcentajes para elegirlos.
- **Valor a la derecha**: lo que queda (`limit - used`), o `Agotado`. Con `used` pero sin `limit` muestra lo usado; sin ninguno, `--`. No se inventa nada: OpenRouter se muestra en USD con los valores que devuelve la API, sin límite mensual ni renovación.
- **Renovación**: `↻ HH:MM`, o con la fecha cuando es otro día. No se muestra cuando la API envía `null`.
- **Act.**: cuándo se observó al proveedor por última vez (`updatedAt`). La línea de abajo es otra cosa: cuándo la placa leyó la API.
- **Estado de la cuenta**: `↻ Desact.` (`stale`), `⚠ Sin sesion` (`auth_required`) o `✕ Sin datos` (`unavailable`), al lado del nombre.
- **MOCK**: en naranja, en las cuentas con `source: mock`. Esos números son de ejemplo, no tu consumo real.
- **Botón BOOT** (al costado de la placa, junto a RESET): muestra la página siguiente en el momento, sin esperar los 10 segundos de la rotación.
- **LED RGB**: sigue a la página en pantalla, con el peor `usageLevel` de sus cupos: `low` verde, `medium` naranja, `high` rojo; apagado cuando no hay nivel o no hay datos. `LED_BRIGHTNESS` en `config.h` regula su brillo (0 lo apaga). Si muestra rojo donde debería ir verde, poné `LED_COLOR_ORDER` en `LED_COLOR_ORDER_GRB`.
- **Errores de lectura**: los últimos datos quedan en pantalla y la línea de abajo dice por qué no se pudo leer (sin Wi-Fi, sin conexión a la API, token rechazado, error HTTP). Tras 150 segundos sin una lectura correcta, todas las cuentas se marcan como desactualizadas. La placa reintenta sola, esperando de 5 segundos a 5 minutos entre intentos.

Los textos se dibujan sin acentos ("Creditos", "dias"): las fuentes incluidas en LVGL no tienen letras acentuadas.

La pantalla admite 16 cuentas y 48 cupos en total; lo que exceda eso no se muestra.

## HTTPS

Con una dirección `https://` el certificado se valida siempre; no hay opción para saltearlo.

- Certificado de una autoridad pública (Let's Encrypt y similares): dejá `API_CA_CERT` vacío. Se usan las autoridades incluidas en el core de ESP32.
- Autoridad privada o certificado autofirmado: pegá el certificado raíz en formato PEM en `API_CA_CERT`.

HTTPS necesita la fecha actual, que la placa obtiene por NTP, así que necesita salida a Internet. En una red hogareña confiable, `http://` con la IP LAN es más simple.

## Problemas frecuentes

| Problema | Solución |
|---|---|
| `lv_conf.h: No such file or directory` | Falta LVGL o no es la del ZIP. Comprobá que la carpeta `lvgl` de `libraries` sea la versión 8.3.10 y contenga `src/lv_conf.h`. Borrá cualquier otra LVGL y copiá de nuevo la carpeta completa del ZIP |
| Errores dentro de archivos de LVGL, o `This firmware needs LVGL 8.3.10` | Está instalada LVGL 9. Quitala y usá la carpeta del ZIP |
| `Sketch too big` / el programa no entra | Elegí **Herramientas → Partition Scheme → Huge APP (3MB No OTA/1MB SPIFFS)** |
| `ArduinoJson.h: No such file or directory` | Instalá ArduinoJson 7.4.3 desde **Herramientas → Administrar bibliotecas** |
| `Missing config.h` | Copiá `config.example.h` a `config.h` en la carpeta del sketch (paso 13) |
| La placa **ESP32C6 Dev Module** no está en la lista | Instalá **esp32 by Espressif Systems** 3.3.12 (paso 4) |
| No aparece ningún puerto | Usá un cable USB-C de datos y otro puerto USB. Si sigue sin aparecer, mantené **BOOT**, presioná y soltá **RESET**, soltá **BOOT**, y mirá de nuevo la lista de puertos |
| La carga falla o queda en `Connecting...` | Poné la placa en modo de descarga con **BOOT** y **RESET** como arriba, subí de nuevo y presioná **RESET** al terminar |
| ¿La carga fue exitosa? | Sí, si la salida muestra `Hash of data verified.` y `Hard resetting via RTS pin...` |
| Pantalla negra después de subir | Revisá la placa (ESP32C6 Dev Module) y el esquema de particiones, y presioná **RESET**. Subí de nuevo el demo oficial: si el demo tampoco muestra nada, el problema no es este firmware. Comprobá que `DISPLAY_BRIGHTNESS` no sea 0 |
| La pantalla dice **Falta configurar** | `WIFI_SSID`, `API_BASE_URL` o `DEVICE_TOKEN` están vacíos en `config.h`. Completalos y subí de nuevo |
| Queda en **Sin Wi-Fi** | Revisá el nombre y la contraseña de la red (importan mayúsculas y minúsculas). La red tiene que ser de 2,4 GHz: la placa no ve redes de 5 GHz |
| **localhost no sirve** | `API_BASE_URL` usa `localhost` o `127.0.0.1`. Usá la IP LAN o el hostname del servidor |
| **Sin conexion a la API** | IP o puerto incorrectos, servidor apagado, o el servidor solo escucha en sí mismo: configurá `HOST_BIND` (ver [despliegue](deployment.es.md)). Repetí los `curl` del paso 14 desde otra computadora. Con `https://`, revisá también el certificado |
| **Token rechazado (401)** | El `DEVICE_TOKEN` de `config.h` no es el del servidor. Copialo de nuevo, sin espacios ni comillas adentro |
| **Respuesta invalida** | La dirección responde, pero no es esta API: revisá el puerto y que la URL no tenga una ruta de más |
| Todo dice **MOCK** | El servidor usa la configuración de ejemplo. Poné `CONFIG_NAME=config.live.json` en el servidor y conectá tus cuentas; ver [configuración](configuration.es.md) |
| Una cuenta muestra **Sin sesion** o **Sin datos** | Viene del servidor, no de la placa: a esa cuenta le falta el login o la credencial. Ver [conectar providers](providers/connect.es.md) |

Para leer los mensajes de la placa, poné **Herramientas → USB CDC On Boot → Enabled**, subí de nuevo y abrí **Herramientas → Monitor serie** a 115200 baudios. Nunca se imprimen contraseñas ni tokens.

## Cómo está armado

- `ai_quota_monitor_vertical.ino` / `ai_quota_monitor_horizontal.ino`: arranque y bucle de dibujo.
- `layout.h`: orientación de la pantalla y tipo de página. Es el único archivo distinto entre los dos sketches; el resto se mantiene igual en ambas carpetas.
- `net.cpp`: Wi-Fi y consultas, en su propia tarea, para que la pantalla siga refrescando durante una consulta. Timeouts de 8 s para conectar y 10 s para leer.
- `usage_parser.cpp`: lee el contrato de `GET /usage` descrito en la [referencia de la API](api.es.md).
- `ui.cpp`: páginas y dibujo con LVGL.
- `logos.cpp`: logos de los providers, generado por `firmware/esp32-c6-lcd-1.47/tools/make_logos.py` a partir de `server/internal/web/static/logos/`. Para cambiar un logo, reemplazá la imagen y corré el script (usa `sips`, de macOS).
- `Display_ST7789.*`, `LVGL_Driver.*`: drivers de pantalla adaptados del demo oficial de Waveshare. Ver [`NOTICE.md`](../firmware/esp32-c6-lcd-1.47/ai_quota_monitor_vertical/NOTICE.md).

El servidor responde `GET /usage` con lo que tiene guardado; leerlo desde la placa nunca dispara una consulta a un proveedor. La placa lee cada 60 segundos.

## Pendiente en hardware

- Probar la reconexión Wi-Fi y ejecuciones largas.
- Probar una dirección `https://` con un certificado real.
- Probar el botón BOOT y los colores del LED RGB en la placa física.
