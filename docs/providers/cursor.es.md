# Cursor

[English](cursor.md) · **Español**

Estado: **experimental**. Usa el endpoint que alimenta el panel de uso de cursor.com. No está documentado por Cursor, puede cambiar sin aviso y no se validó todavía con una cuenta real; los tests usan respuestas con la forma reportada por terceros.

No requiere ser administrador: lee solo el uso de la cuenta que inició sesión.

## Qué consulta

`GET https://cursor.com/api/usage-summary`, con la cookie `WorkosCursorSessionToken`. No envía prompts ni consume cupo.

## Configuración

1. Abrí <https://cursor.com/dashboard> con la cuenta correcta.
2. DevTools → Application → Cookies → `https://cursor.com` → copiá el valor de `WorkosCursorSessionToken`.
3. Pegalo en `server/.env`:

```env
CURSOR_WORK_SESSION=valor_de_la_cookie
```

4. Declará la cuenta en `server/configs/`:

```json
{
  "id": "cursor-work",
  "name": "Cursor trabajo",
  "provider": "cursor",
  "scope": "work",
  "source": "direct",
  "credentialEnv": "CURSOR_WORK_SESSION"
}
```

5. Reiniciá el servidor y consultá `GET /usage`.

La cookie es una credencial: da acceso a tu cuenta de Cursor. No la subas ni la compartas. Para otra cuenta, agregá otra entrada con otra variable; en Docker Compose también hay que pasarla en `compose.yaml`.

## Cupos devueltos

Todos en `percent`, con `resetsAt` igual al fin del ciclo de facturación:

| ID | Significado | Origen |
|---|---|---|
| `included` | Uso incluido en el plan | `totalPercentUsed`, o `used/limit` de `individualUsage.plan` u `overall` |
| `auto` | Porción de modelos Auto | `autoPercentUsed`, si existe |
| `api` | Porción de modelos elegidos a mano | `apiPercentUsed`, si existe |
| `on-demand` | Gasto on-demand propio contra su tope | `individualUsage.onDemand`, si está habilitado y tiene límite |
| `team-on-demand` | Gasto on-demand del equipo contra su tope | `teamUsage.onDemand`, si está habilitado y tiene límite |

Solo se informan medidores con límite conocido. Se usan porcentajes porque la unidad de `used`/`limit` no está documentada (pedidos o centavos según el plan); no se convierten a USD.

## Límites conocidos

- La sesión vence y no se renueva sola. Cuando pasa, la cuenta queda en `auth_required`: copiá la cookie de nuevo y reiniciá.
- Algunas cuentas Enterprise no exponen ningún medidor con límite; en ese caso la cuenta queda `unavailable`.
- Las políticas de la organización pueden limitar lo que muestra el dashboard.
