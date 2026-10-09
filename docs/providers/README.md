# Proveedores

[English](README.en.md) · **Español**

Para conectar tus cuentas, ver la [guía paso a paso](conectar.md).

La implementación específica de cada servicio vive en `server/internal/providers/<proveedor>/`:

- `codex/`: cliente de Codex App Server, autenticación, sesiones y transformación de cupos.
- `openrouter/`: cliente HTTP y consulta de créditos.
- `cursor/`: cliente HTTP del resumen de uso del dashboard.
- `claude/`: login OAuth, renovación de sesión y lectura de límites de la suscripción.

Los comandos y la API HTTP usan estos paquetes; los próximos proveedores deben seguir la misma estructura.

## OpenRouter — cliente implementado

`source: direct`, `provider: openrouter`, `credentialEnv` apunta a una variable del servidor. Consulta `/api/v1/credits` con Bearer y management key. Devuelve total comprado y total usado, en USD. El saldo es la diferencia; no es una asignación mensual. 401/403 producen auth_required. No se probaron credenciales reales durante la creación del proyecto.

[Documentación oficial](https://openrouter.ai/docs/api/api-reference/credits/get-remaining-credits)

## Codex / Work — integrado en servidor

Consultar cupos propios con `account/rateLimits/read` de Codex App Server, usando autenticación correspondiente. Puede haber varios buckets; conservar duración y renovación reales en vez de asumir siempre 5h/semanal. Validar sesiones aisladas para cuentas personales/laborales sin cambiar la sesión principal. La API inicia login por código mediante HTTP y consulta los cupos directamente en el homelab. Ver [setup de Codex](codex.md).

[Documentación oficial](https://learn.chatgpt.com/docs/app-server)

Los datos no representan todos los cupos del chat normal de ChatGPT.

## Claude — integrado en servidor, experimental

`source: direct`, `provider: anthropic`. Login OAuth por HTTP y lectura de las ventanas de 5h y 7d desde un endpoint no documentado. Es la suscripción, no gasto de API. **Los términos de Anthropic no permiten este uso de las credenciales**; leer la advertencia en [setup de Claude](claude-subscription.md) antes de activarlo. No se validó con una cuenta real.

## Cursor — cliente implementado, experimental

`source: direct`, `provider: cursor`, `credentialEnv` apunta a la cookie de sesión de cursor.com. Consulta el endpoint no documentado `/api/usage-summary` con permisos de usuario normal, sin Admin API. Devuelve porcentajes con la renovación del ciclo de facturación. No se validó con una cuenta real. Ver [setup de Cursor](cursor.md).

## Kiro — investigación pendiente

El usuario objetivo no tiene acceso admin. No asumir que existe una API pública Kiro de saldo. Validar qué datos permite leer la propia sesión. Si se utiliza una interfaz privada del dashboard, documentar dependencia y fragilidad, mantener permisos normales de usuario y marcar experimental. Nunca inventar cupos a partir de gastos o tokens sin un límite conocido.

- [Kiro FAQ](https://kiro.dev/faq/)

Los ejemplos mock son exclusivamente para diseño. No implementan autenticación ni scraping de estos servicios.
