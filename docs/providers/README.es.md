# Proveedores

[English](README.md) · **Español**

Para conectar tus cuentas, ver la [guía paso a paso](connect.es.md).

La implementación específica de cada servicio vive en `server/internal/providers/<proveedor>/`:

- `codex/`: cliente de Codex App Server, autenticación, sesiones y transformación de cupos.
- `openrouter/`: cliente HTTP y consulta de créditos.
- `cursor/`: cliente HTTP del resumen de uso del dashboard.
- `claude/`: login OAuth, renovación de sesión y lectura de límites de la suscripción.

Los comandos y la API HTTP usan estos paquetes; los próximos proveedores deben seguir la misma estructura. Ver [cómo agregar un proveedor](../../CONTRIBUTING.es.md#agregar-un-proveedor).

## OpenRouter — cliente implementado

`source: direct`, `provider: openrouter`, `credentialEnv` apunta a una variable del servidor. Consulta `/api/v1/credits` con Bearer y management key. Devuelve total comprado y total usado, en USD. El saldo es la diferencia; no es una asignación mensual. 401/403 producen auth_required. No se probaron credenciales reales durante la creación del proyecto.

[Documentación oficial](https://openrouter.ai/docs/api/api-reference/credits/get-remaining-credits)

## Codex / Work — integrado en servidor

Consultar cupos propios con `account/rateLimits/read` de Codex App Server, usando autenticación correspondiente. Puede haber varios buckets; conservar duración y renovación reales en vez de asumir siempre 5h/semanal. Validar sesiones aisladas para cuentas personales/laborales sin cambiar la sesión principal. La API inicia login por código mediante HTTP y consulta los cupos directamente en el homelab. Ver [setup de Codex](codex.es.md).

[Documentación oficial](https://learn.chatgpt.com/docs/app-server)

Los datos no representan todos los cupos del chat normal de ChatGPT.

## Claude — integrado en servidor, experimental

`source: direct`, `provider: anthropic`. Login OAuth por HTTP y lectura de las ventanas de 5h y 7d desde un endpoint no documentado. Es la suscripción, no gasto de API. **Los términos de Anthropic no permiten este uso de las credenciales**; leer la advertencia en [setup de Claude](claude-subscription.es.md) antes de activarlo. No se validó con una cuenta real.

## Cursor — cliente implementado, experimental

`source: direct`, `provider: cursor`, `credentialEnv` apunta a la cookie de sesión de cursor.com. Consulta el endpoint no documentado `/api/usage-summary` con permisos de usuario normal, sin Admin API. Devuelve porcentajes con la renovación del ciclo de facturación. No se validó con una cuenta real. Ver [setup de Cursor](cursor.es.md).

## Kiro — evaluado, sin implementar

Solo existe como cuenta de ejemplo (`source: mock`). No hay una API documentada de saldo de Kiro, pero el IDE de Kiro y el comando `/usage` de `kiro-cli` leen el consumo del propio usuario con una operación no documentada, `GetUsageLimits` (`GET /getUsageLimits`, access token Bearer y `profileArn`), con permisos de usuario normal. Esto sale de leer el cliente instalado; no se hizo ninguna llamada real.

- Respuesta: nombre del plan, `nextDateReset` y, por recurso, los créditos usados y el límite (`usageBreakdownList`), más créditos de prueba o bonos y excedentes.
- Mapeo: un cupo `credits` con usado, límite y `resetsAt`, como la cuenta de ejemplo `kiro-work`. Omitir el cupo cuando el plan no tiene límite y dejar afuera los excedentes: nunca inventar cupos a partir de gastos o tokens sin un límite conocido.
- Sesión: el access token dura poco y se renueva con un refresh token; un cliente `direct` necesitaría su propio login por código de dispositivo, como Codex y Claude.
- Duda abierta: las cuentas administradas por una organización (IAM Identity Center) pueden no recibir los números; el cliente contempla un error de "función no soportada" para esta llamada.

Si se implementa, marcarlo experimental y documentar la dependencia del endpoint no documentado.

- [Kiro FAQ](https://kiro.dev/faq/)

Los ejemplos mock son exclusivamente para diseño. No implementan autenticación ni scraping de estos servicios.
