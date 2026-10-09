# Carcasa — pendiente de diseño

[English](README.md) · **Español**

Se publicarán fuente editable en `source/` y STL en `stl/` cuando el modelo esté diseñado y probado. No se incluyen STL ficticios ni dimensiones asumidas.

Requisitos: acceso USB-C, BOOT/RESET, ventilación, pantalla sin presión, soporte de escritorio y método de fijación. Medir placa, conectores, tolerancias y revisar planos oficiales. Imprimir una prueba de encastre antes de publicar versión lista para replicar. Documentar material, orientación, capas, soportes y tornillos según diseño final.

## Estado y próximos pasos

Las carpetas `source/` y `stl/` están vacías: todavía no hay ningún archivo para imprimir.

1. Medir la placa real y revisar los planos del fabricante. Ver [hardware](../docs/hardware.es.md).
2. Diseñar un modelo paramétrico (OpenSCAD o un CAD equivalente) y guardarlo en `source/`.
3. Imprimir una prueba de encastre y ajustar tolerancias.
4. Publicar los STL en `stl/` con los parámetros de impresión usados.
