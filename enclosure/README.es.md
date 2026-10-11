# Carcasa

[English](README.md) · **Español**

![Vista previa de la carcasa: frente con la pantalla horizontal, lateral con el USB-C y parte trasera con la tapa a presión](preview.png)

Carcasa de escritorio para la [Waveshare ESP32-C6-LCD-1.47](../docs/hardware.es.md), versión 3: dos piezas, pantalla horizontal inclinada 12° hacia atrás y tapa a presión, sin tornillos. Va con el sketch [`ai_quota_monitor_horizontal`](../firmware/esp32-c6-lcd-1.47/README.es.md).

Estado: diseñada e impresa en una Bambu Lab A1 mini. La imagen es un render generado por el script del modelo, no una foto.

## Archivos

| Archivo | Para qué |
|---|---|
| [`stl/claude-esp32-v6.3mf`](stl/claude-esp32-v6.3mf) | Las dos piezas juntas, para abrir en Bambu Studio u otro laminador |
| [`stl/carcasa.stl`](stl/carcasa.stl) | Cuerpo: 66 × 39,5 × 28 mm |
| [`stl/tapa.stl`](stl/tapa.stl) | Tapa trasera con las trabas: 51 × 31 × 12 mm |
| [`source/generar_modelo.py`](source/generar_modelo.py) | Modelo paramétrico: genera los STL, la vista previa y `validacion.json` |
| [`source/empaquetar.py`](source/empaquetar.py) | Arma el 3MF a partir de los STL |
| [`source/validacion.json`](source/validacion.json) | Medidas y chequeos geométricos de la última generación |

Los STL están en milímetros y a escala 100 %.

## Diseño

- **Apoyo**: cuatro patas y un pie trasero que comparten el plano de apoyo; la pantalla queda inclinada 12°.
- **Tapa**: dos lengüetas flexibles con gancho, impresas junto con la tapa. No lleva tornillos ni piezas adicionales.
- **USB-C a la derecha**: abertura de 15 × 8,8 mm con marco reforzado. El borde trasero tiene 2,4 mm.
- **BOOT y RESET**: accesibles por arriba y por abajo, cerca del extremo del USB.
- **LED RGB**: ventanas abiertas en las paredes; no se imprimen difusores.
- **Ventilación**: tres ranuras en la tapa.
- **Placa**: el módulo se toma como 36,37 × 20,32 × 7,60 mm, según el [plano de Waveshare](https://docs.waveshare.com/ESP32-C6-LCD-1.47), con 0,45 mm de holgura por lado. La ranura microSD queda accesible con la tapa quitada.

## Impresión

1. Abrí `stl/claude-esp32-v6.3mf` en el laminador, como geometría. Contiene solo la carcasa y la tapa: elegí impresora, filamento y perfil.
2. Mantené la escala en 100 %.
3. La carcasa se imprime con el frente hacia la cama; la tapa, con la cara exterior plana hacia la cama.
4. Revisá en el laminado los puentes de la abertura del USB y las rampas chicas de los ganchos.
5. Evitá soportes que unan o dañen las lengüetas.

La carcasa se imprimió en una **Bambu Lab A1 mini**; el 3MF se abrió y se editó en **Bambu Studio** antes de imprimir. El material, la altura de capa y el perfil usados no están documentados todavía.

## Armado

1. Probá primero sin cable: apoyá el módulo en la cavidad con el USB-C a la derecha, sin presionar el vidrio de la pantalla.
2. Alineá la tapa y presioná suavemente hasta que enganchen las dos trabas. Si pide fuerza, sacala y revisá la holgura y el acabado de las lengüetas.
3. Conectá el cable y comprobá que no desplace el módulo.

Para abrir, presioná las dos lengüetas hacia el centro por las ventanas de arriba y de abajo, y tirá hacia atrás de la pestaña lateral de la tapa.

## Regenerar el modelo

```sh
pip install manifold3d numpy trimesh matplotlib pillow
python3 enclosure/source/generar_modelo.py
python3 enclosure/source/empaquetar.py
```

El primer script reescribe `stl/carcasa.stl`, `stl/tapa.stl`, `preview.png` y `source/validacion.json`; el segundo, `stl/claude-esp32-v6.3mf`. La vista previa usa la fuente Arial de macOS.

## Qué se verificó

- Geometría: cada pieza es un sólido cerrado, y Bambu Studio reconoce dos sólidos válidos en el 3MF.
- Sin interferencia entre carcasa, tapa, módulo y un cable de referencia de 12 × 6 mm en posición de cierre.
- Los ganchos retienen la tapa, y la tapa sale flexionando 0,6 mm cada traba.
- Patas y pie comparten el plano de apoyo en posición de uso.

Estos chequeos son geométricos: no simulan esfuerzos ni fatiga. La durabilidad de las trabas no se midió.
