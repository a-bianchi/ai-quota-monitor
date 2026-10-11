from pathlib import Path
import xml.etree.ElementTree as E
import zipfile
import trimesh
p=Path(__file__).resolve().parent
ns='http://schemas.microsoft.com/3dmanufacturing/core/2015/02'
E.register_namespace('',ns)
def el(parent,tag,**kw): return E.SubElement(parent,'{'+ns+'}'+tag,kw)
root=E.Element('{'+ns+'}model',{'unit':'millimeter','{http://www.w3.org/XML/1998/namespace}lang':'es-AR'})
el(root,'metadata',name='Title').text='ESP32-C6 LCD 1.47 - v6 USB lado opuesto respecto de la pieza v5 impresa - tapa corrediza - medidas confirmadas - inclinacion 12 grados'
resources=el(root,'resources');build=el(root,'build')
for i,(name,(x,y)) in enumerate(zip(['carcasa','tapa'],[(50,50),(120,50)]),1):
 t=trimesh.load_mesh(p/(name+'.stl'))
 ob=el(resources,'object',id=str(i),type='model',name=name)
 me=el(ob,'mesh');vs=el(me,'vertices');ts=el(me,'triangles')
 for v in t.vertices: el(vs,'vertex',x=f'{v[0]:.6f}',y=f'{v[1]:.6f}',z=f'{v[2]:.6f}')
 for tri in t.faces: el(ts,'triangle',v1=str(tri[0]),v2=str(tri[1]),v3=str(tri[2]))
 el(build,'item',objectid=str(i),transform=f'1 0 0 0 1 0 0 0 1 {x} {y} 0')
with zipfile.ZipFile(p/'claude-esp32-v6_USB_LADO_OPUESTO.3mf','w',zipfile.ZIP_DEFLATED) as z:
 z.writestr('[Content_Types].xml','<?xml version="1.0" encoding="UTF-8"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="model" ContentType="application/vnd.ms-package.3dmanufacturing-3dmodel+xml"/></Types>')
 z.writestr('_rels/.rels','<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Target="/3D/3dmodel.model" Id="rel0" Type="http://schemas.microsoft.com/3dmanufacturing/2013/01/3dmodel"/></Relationships>')
 z.writestr('3D/3dmodel.model',E.tostring(root,encoding='utf-8',xml_declaration=True))
with zipfile.ZipFile(p/'claude-esp32-v6_USB_LADO_OPUESTO.zip','w',zipfile.ZIP_DEFLATED) as z:
 for f in p.iterdir():
  if f.suffix in ['.stl','.3mf','.txt','.png','.json','.py']: z.write(f,f.name)
print('3MF y ZIP v6 generados')
