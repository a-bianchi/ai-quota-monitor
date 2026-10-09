# Claude (suscripción Pro/Max)

[English](claude-subscription.md) · **Español**

Muestra los límites de la suscripción de Claude: ventana de 5 horas, ventana de 7 días y, si existen, las semanales por modelo. No son créditos de la API de Anthropic.

## Advertencia: leer antes de usar

Estado: **experimental y no permitido por los términos de Anthropic.**

Anthropic no publica una API para leer los límites de una suscripción. Este provider inicia sesión con el cliente OAuth público de Claude Code y guarda los tokens en el servidor. La [página legal de Claude Code](https://code.claude.com/docs/en/legal-and-compliance) dice que los desarrolladores de terceros no pueden ofrecer login de Claude.ai en sus aplicaciones ni recolectar o guardar credenciales o tokens de sesión de Claude.ai, y que Anthropic puede aplicar esas restricciones sin aviso.

Usarlo es tu decisión y el riesgo es sobre tu cuenta. Si no querés asumirlo, borrá la cuenta `anthropic` de tu configuración: el resto del monitor funciona igual.

Además, el endpoint de uso no está documentado y puede cambiar o dejar de funcionar. La integración se probó solo con endpoints simulados.

## Qué consulta

- Login: autorización OAuth con PKCE, pidiendo solo el permiso `user:profile`. La sesión no sirve para enviar prompts.
- Uso: `GET https://api.anthropic.com/api/oauth/usage`. No envía prompts ni consume cupo.

## Configuración

En `server/configs/config.live.json`:

```json
{"id":"claude-personal","name":"Claude personal","provider":"anthropic","scope":"personal","source":"direct"}
```

Requiere `COLLECTOR_TOKEN`. Reiniciá el servidor después de cambiar la configuración.

## Login HTTP

1. Iniciar:

   ```sh
   curl -X POST http://TU_SERVIDOR:8080/accounts/claude-personal/claude/login \
     -H 'Authorization: Bearer TU_COLLECTOR_TOKEN'
   ```

   Respuesta `202`:

   ```json
   {"status":"pending","verificationUrl":"https://claude.com/cai/oauth/authorize?...","expiresAt":"2026-10-09T16:00:00Z"}
   ```

2. Abrí `verificationUrl` en tu navegador, iniciá sesión con la cuenta correcta y autorizá. La página muestra un código con la forma `codigo#estado`.

3. Enviar el código completo:

   ```sh
   curl -X PUT http://TU_SERVIDOR:8080/accounts/claude-personal/claude/login \
     -H 'Authorization: Bearer TU_COLLECTOR_TOKEN' \
     -H 'Content-Type: application/json' \
     -d '{"code":"codigo#estado"}'
   ```

   Respuesta `200` con `{"status":"authenticated"}`. El servidor consulta los límites enseguida.

4. Estado, cuando quieras: `GET /accounts/claude-personal/claude/login`.

Errores del paso 3: `409` si no hay login en curso o pasaron los 15 minutos, `400` si el código fue rechazado, `502` si falló la comunicación con Claude. En los tres casos se empieza de nuevo con el paso 1.

Estados: `pending`, `authenticated`, `auth_required`, `unknown`. Después de reiniciar el estado es `unknown` hasta la primera lectura; no significa que la sesión se perdió.

## Cupos devueltos

| ID | Significado |
|---|---|
| `five-hour` | Ventana de 5 horas |
| `seven-day` | Ventana de 7 días, todos los modelos |
| `seven-day-opus`, `seven-day-sonnet` | Semanales por modelo, si la respuesta los incluye |

Todos en `percent`, con `resetsAt` cuando Claude lo informa.

## Persistencia y límites

- La sesión se guarda en `DATA_DIR/claude/<account-id>/credentials.json`; en Docker, dentro del volumen `usage-data`. Es un secreto: no exponerlo ni subirlo a Git.
- El servidor renueva la sesión solo. Si la renovación es rechazada, la cuenta pasa a `auth_required` y hay que repetir el login.
- El ID de la cuenta no verifica con qué cuenta iniciaste sesión en el navegador.
- No se confirmó con una cuenta real que el permiso `user:profile` alcance para leer el uso.
