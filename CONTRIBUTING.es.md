# Cómo contribuir

[English](CONTRIBUTING.md) · **Español**

Gracias por tu interés en el proyecto. Esta guía explica cómo ejecutarlo, probarlo y proponer cambios.

## Ejecutar el proyecto

Requisitos: Git, Go 1.22 o posterior, `make` y `openssl`. Desde la raíz del repo:

```sh
cp server/.env.example server/.env
# Pegá en DEVICE_TOKEN la salida de: openssl rand -hex 32
make run
```

El servidor arranca en `http://localhost:8080` con datos de ejemplo, sin ninguna cuenta real. Para el panel web, creá un usuario con `make panel-password`. Más detalle en [instalación local](docs/local-setup.es.md).

No hace falta ninguna credencial de proveedor para desarrollar: el modo de ejemplo y los tests usan datos y servidores simulados.

## Correr las pruebas

```sh
make test
cd server && go vet ./...
```

`make test` ejecuta `go test -race ./...` dentro de `server/`. El workflow de GitHub Actions (`.github/workflows/go.yml`) corre esos mismos dos pasos en cada push y pull request.

Antes de enviar un cambio, formateá el código con `gofmt`:

```sh
cd server && gofmt -l .
```

El comando no debe listar ningún archivo.

## Estructura

El código vive en `server/`; [arquitectura](docs/architecture.es.md) describe cada paquete. El servidor usa solo la biblioteca estándar de Go: no agregues dependencias externas sin discutirlo antes en un issue.

## Agregar un proveedor

Cada integración vive en su propio paquete, `server/internal/providers/<proveedor>/`. Tomá `openrouter/` o `cursor/` como referencia: son los más simples.

1. **Creá el paquete** con un cliente que devuelva un `usage.Report`:

   ```go
   func (p Provider) Collect(ctx context.Context, credential string) (usage.Report, error)
   ```

   - Convertí la respuesta del proveedor a cupos (`usage.Quota`) con `unit` en `percent`, `credits`, `usd` o `tokens`.
   - Validá el resultado con `usage.ValidateReport` antes de devolverlo.
   - Exportá un error `ErrAuth` y devolvelo cuando la credencial falte, sea rechazada o haya vencido.
   - Usá timeouts, no registres credenciales ni cuerpos de respuesta, y no envíes prompts ni consumas cupo.
   - Informá solo medidores con un límite conocido: no inventes cupos a partir de gastos sin límite.

2. **Registrá el proveedor** en `server/internal/config/config.go`: agregalo a la lista de `provider` válidos y, si admite `source: direct`, a ese `switch`, indicando si requiere `credentialEnv`.

3. **Conectalo al ciclo de consulta** en `server/cmd/server/main.go`: agregá su caso al `switch a.Provider` de `collect` y sumá su `ErrAuth` a la condición que produce `auth_required`.

4. **Agregá los textos del panel** en `server/internal/web/static/i18n/es.json` y `en.json`: el nombre en `provider.<proveedor>` y, si usa una credencial, su etiqueta en `settings.credential.field.<proveedor>`. Los nombres de cupos se escriben en español; su traducción al inglés va en `server/internal/httpapi/localize.go`.

5. **Escribí tests** con `httptest`, cubriendo la respuesta correcta, el error de autenticación y las respuestas incompletas. No uses credenciales reales ni llames al servicio real en los tests.

6. **Documentalo** en `docs/providers/<proveedor>.md` y `docs/providers/<proveedor>.es.md`: autenticación, endpoint consultado, unidades, cupos devueltos, renovación y límites conocidos. Indicá con claridad si el endpoint es oficial y si se validó con una cuenta real. Actualizá las tablas de `docs/providers/README.md`, `docs/providers/connect.md` y el README principal, en los dos idiomas.

7. **Si usa una variable de entorno**, agregala vacía en `server/.env.example` y en la sección `environment` de `compose.yaml`.

Si el proveedor necesita un login interactivo, como Codex o Claude, mirá esos paquetes y sus rutas en `server/internal/httpapi/api.go`. Conviene abrir un issue antes para acordar el diseño.

## Convenciones

- **Cambios chicos y enfocados**: un pull request por tema.
- **Mensajes de commit** cortos, con un prefijo que indique el tipo: `feat:`, `fix:`, `docs:`, `test:`, `refactor:`.
- **Código** en el estilo del archivo que tocás, formateado con `gofmt`. Identificadores, comentarios y mensajes de error en inglés.
- **Tests** para todo cambio de comportamiento.
- **Documentación** en inglés (`archivo.md`) y en español (`archivo.es.md`), con los enlaces entre ambas versiones al comienzo. Si cambiás una, actualizá la otra.
- **Sin funciones inventadas**: documentá lo que funciona y marcá como pendiente o experimental lo que no se validó.

## Pull requests

1. Hacé un fork y creá una rama desde `main`.
2. Aplicá el cambio con sus tests y su documentación.
3. Corré `make test` y `cd server && go vet ./...`.
4. Abrí el pull request explicando qué cambia, por qué y cómo lo probaste.

## Seguridad y datos privados

- No incluyas tokens, keys, cookies, archivos de sesión, emails ni información de trabajo en commits, issues, capturas o logs.
- `server/.env`, `server/data/` y `server/configs/config.local.json` están en `.gitignore`: no fuerces su inclusión.
- Las capturas deben usar el modo de ejemplo.
- Si encontrás un problema de seguridad, no publiques los detalles en un issue: contactá primero a la persona que mantiene el repositorio.

## Licencia

Al contribuir aceptás que tu aporte se publique bajo la [licencia MIT](LICENSE) del proyecto.
