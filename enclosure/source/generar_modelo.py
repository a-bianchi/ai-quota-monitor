"""Modelo de prueba, mm. pip install manifold3d numpy trimesh matplotlib pillow.
Plano: https://docs.waveshare.com/ESP32-C6-LCD-1.47
Cotas de botones, USB y difusores incluyen margen y requieren prueba fisica.
"""
from pathlib import Path
import json, math, zipfile
import numpy as np
import manifold3d as m
import trimesh
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt
from mpl_toolkits.mplot3d.art3d import Poly3DCollection

OUT = Path(__file__).resolve().parent
ROOT = OUT.parent # enclosure/
STL = ROOT/'stl'
PCB_W, PCB_H, PCB_D = 36.37, 20.32, 7.60
CLEAR = 0.45 # reduced lateral play; no removable diffuser flanges
CW, CH = PCB_W+2*CLEAR, PCB_H+2*CLEAR
W,H,D = 48.,31.,16.
FRONT = 3.2 # deeper bezel seat, recesses module 1.2 mm more than v2
LCD_W,LCD_H,LCD_X = 32.95,17.99,-0.965
LIGHT_Z,LIGHT_H = 5.2,1.6
TILT_DEG = 12.0
SLOPE = math.tan(math.radians(TILT_DEG))
SHEAR = [[1,0,0,0],[0,1,SLOPE,0],[0,0,1,0]]

# Box corners in local X=width, Y=height, Z=depth from front.
def box(x0,x1,y0,y1,z0,z1):
    return m.Manifold.cube((x1-x0,y1-y0,z1-z0)).translate((x0,y0,z0))
def cyl(x,y,z,r,h):
    return m.Manifold.cylinder(h,r,r,48).translate((x,y,z))
def union(items):
    result=items[0]
    for item in items[1:]: result=result+item
    return result

shape=box(-W/2,W/2,-H/2,H/2,0,D)
for side in [-1,1]:
    # Pixel shoulder and hanging forearm; all begin on build plate.
    shape += box(side*27-4,side*27+4,-1,9,0,D)
    shape += box(side*29-3,side*29+3,-9,5,0,D)
# Reinforced USB surround extends 2 mm beyond right arm.
shape += box(30.5,34,-9,9,2.5,D)
for x in [-18,-6,6,18]:
    shape += box(x-3,x+3,-24,-14.5,0,D)
# Rear foot joins all legs at their base; assembled support depth 28 mm.
# Rear base follows the sloped contact plane while retaining 3 mm thickness.
shape += box(-24,24,-24,-21,12,28).transform(SHEAR)
# Clearance above rear foot permits withdrawing lid straight backward.
shape -= box(-24.1,24.1,-15.8,-14,23,28.1)
# Remove material below the tabletop contact plane from all four legs.
shape -= box(-100,100,-100,-24,-10,100).transform(SHEAR)
shape -= box(-CW/2,CW/2,-CH/2,CH/2,FRONT,D+1)
shape -= box(LCD_X-LCD_W/2,LCD_X+LCD_W/2,-LCD_H/2,LCD_H/2,-1,FRONT+0.1)
# BOOT/RST on long edges, near USB end. Broad tool/fingertip access.
BUTTON_X = PCB_W/2-7.04
for s in [-1,1]:
    a,b = sorted([s*(CH/2-0.1),s*(H/2+1)])
    shape -= box(BUTTON_X-3.5,BUTTON_X+3.5,a,b,7.7,14.0)
# USB-C at right. Fits plug envelope provisionally; SD accessible with lid off.
USB_WIDTH=15.0
USB_HEIGHT=8.8 # 6 mm cable thickness plus clearance and position margin
USB_Z=FRONT+6.0
shape -= box(CW/2-0.1,35,-USB_WIDTH/2,USB_WIDTH/2,
             USB_Z-USB_HEIGHT/2,USB_Z+USB_HEIGHT/2)
# Light channels through the walls, backed by removable translucent inserts.
for s in [-1,1]:
    a,b=sorted([s*(CH/2-0.1),s*(H/2+1)])
    shape -= box(-14,13,a,b,LIGHT_Z,LIGHT_Z+LIGHT_H)
    a,b=sorted([s*(CW/2-0.1),s*33])
    shape -= box(a,b,-4.5,4.5,LIGHT_Z,LIGHT_Z+LIGHT_H)
# Two captive snap beams on the cover, accessible from top and bottom.
# Beam pockets are outside PCB envelope and away from USB opening.
for sign in [-1,1]:
    lo,hi=sorted([sign*11.5,sign*13.55])
    shape -= box(-4,4,lo,hi,6.2,D+1)
    # Release windows double as hook seats. Fingers squeeze toward screen center.
    lo,hi=sorted([sign*11.5,sign*16.5])
    shape -= box(-4.4,4.4,lo,hi,3.6,7.65)

# Cover printed exterior face down. Interior lip positions cover.
T=2.4
lid=box(-24,24,-15.5,15.5,0,T)
lip=box(-CW/2+0.25,CW/2-0.25,-CH/2+0.25,CH/2-0.25,T,T+1.2)
lip -= box(-CW/2+1.45,CW/2-1.45,-CH/2+1.45,CH/2-1.45,T-0.1,T+1.3)
lid += lip
# Clip beams: 9.6 mm long, 1.2 mm thick, 6 mm wide.
# Hook ramp bends the beam inward on insertion; square shoulder retains cover.
for sign in [-1,1]:
    lo,hi=sorted([sign*12.1,sign*13.3])
    lid += box(-3,3,lo,hi,T-0.1,T+9.6)
    hook=m.CrossSection([[(13.15,T+8.7),(13.85,T+8.7),
                         (13.85,T+9.0),(13.15,T+9.6)]]).extrude(6)
    hook=hook.transform([[0,0,1,-3],[1,0,0,0],[0,1,0,0]])
    if sign == -1: hook=hook.mirror((0,1,0))
    lid += hook
# Rear pull grip; permits withdrawing cover after squeezing both clips.
lid += box(-27,-23,-5,5,0,T)

# Rear ventilation, away from board mounting hole locations.
for x in [-6,0,6]: lid -= box(x-1,x+1,-6,6,-0.1,T+0.1)

# Four rear stops aligned to original M2 standoffs in official drawing.
# They stop 0.2 mm behind nominal standoffs rather than preload the LCD.
BOARD_FIX=[(x,y) for x in [-PCB_W/2+1.97,PCB_W/2-2.4] for y in [-6.64,6.64]]
stop_height=D-(FRONT+PCB_D+0.2)
for x,y in BOARD_FIX:
    lid += cyl(x,-y,T-0.1,1.7,stop_height+0.1)

def mesh(man):
    data=man.to_mesh()
    return trimesh.Trimesh(vertices=np.array(data.vert_properties)[:,:3],faces=np.array(data.tri_verts),process=True)
parts={'carcasa':shape,'tapa':lid}
report={'inclinacion_grados':TILT_DEG}
for name,man in parts.items():
    t=mesh(man)
    assert t.is_watertight and t.is_volume, name
    assert len(man.decompose())==1, name
    t.apply_translation([0,0,-t.bounds[0,2]])
    t.export(STL/f'{name}.stl')
    report[name]={'dimensiones_mm':t.extents.round(3).tolist(),'triangulos':len(t.faces),'volumen_mm3':round(float(t.volume),2),'cerrada':bool(t.is_watertight),'solidos':len(man.decompose())}

# Installed meshes for preview. Assembly preview is not a print file.
cover=lid.rotate((180,0,0)).translate((0,0,D+T))
assert (shape ^ cover).volume()<0.001
board_envelope=box(-PCB_W/2,PCB_W/2,-PCB_H/2,PCB_H/2,FRONT+0.01,FRONT+PCB_D)
assert (shape ^ board_envelope).volume()<0.001
assert (cover ^ board_envelope).volume()<0.001
# Verify hook retention and a geometric withdrawal path with 0.6 mm clip squeeze.
locked_overlap=(shape ^ cover.translate((0,0,0.6))).volume()
assert locked_overlap>0.05
released=cover
for sign in [-1,1]:
    lo,hi=sorted([sign*12.0,sign*14.0])
    clip=cover ^ box(-3.1,3.1,lo,hi,5.5,D-0.01)
    released=(released-clip)+clip.translate((0,-sign*0.6,0))
withdraw=[]
for step in [0,0.3,0.6,1,2,4,8,12]:
    collision=(shape ^ released.translate((0,0,step))).volume()
    
    assert collision<0.001,(step,collision)
    withdraw.append(round(collision,6))
cable_reference=box(PCB_W/2,36,-6,6,USB_Z-3,USB_Z+3)
assert (shape ^ cable_reference).volume()<0.001
assert (cover ^ cable_reference).volume()<0.001
report['ensamble']={'interferencia_carcasa_tapa_mm3':round((shape^cover).volume(),5),
                   'retencion_al_tirar_sin_liberar_mm3':round(locked_overlap,4),
                   'recorrido_liberado_interferencias_mm3':withdraw,
                   'espesor_borde_trasero_usb_mm':round(D-USB_Z-USB_HEIGHT/2,2),
                   'hueco_usb_mm':[USB_WIDTH,USB_HEIGHT],
                   'cable_referencia_mm':[12,6],
                   'interferencia_cable_mm3':round((shape^cable_reference).volume(),5)}
# Check the real assembled contact plane, including every foot and rear base.
vertices=mesh(shape).vertices
angle=math.radians(TILT_DEG)
heights=math.cos(angle)*(vertices[:,1]+24)-math.sin(angle)*vertices[:,2]
assert heights.min()>-1e-5
contact=vertices[np.abs(heights)<1e-5]
assert len(contact)>=12
report['apoyo']={'altura_minima_mm':round(float(heights.min()),6),
                 'vertices_coplanares':len(contact),'profundidad_apoyo_mm':round(28/math.cos(angle),3)}
(OUT/'validacion.json').write_text(json.dumps(report,indent=2))

# Render geometry directly; opaque orange shell and milk-white diffuser pieces.
def upright(t):
    verts=t.vertices.copy()
    # The table plane becomes Z=0; screen normal points upward by TILT_DEG.
    c=math.cos(math.radians(TILT_DEG));s=math.sin(math.radians(TILT_DEG))
    h=verts[:,1]+24
    return np.column_stack((verts[:,0],c*verts[:,2]+s*h,c*h-s*verts[:,2]))

# Orthographic z-buffer renderer avoids painter-order artifacts on CAD cavities.
from PIL import Image, ImageDraw, ImageFont
from matplotlib.colors import to_rgb
CANVAS=Image.new('RGB',(1920,920),'#ece9e4')
def font(size):
    return ImageFont.truetype('/System/Library/Fonts/Supplemental/Arial.ttf',size)
draw=ImageDraw.Draw(CANVAS)
draw.text((960,40),'ESP32-C6 LCD 1.47 · v3 sin tornillos · 12°',font=font(32),fill='#292929',anchor='mm')
scene=[(shape,'#ed6b32'),(cover,'#d95023')]+[(box(LCD_X-LCD_W/2,LCD_X+LCD_W/2,-LCD_H/2,LCD_H/2,FRONT+0.12,FRONT+0.4),'#15232c')]
for view,(cam,title) in enumerate([([0.5,-0.85,0.35],'Frente · pantalla horizontal'),([1,0,0.12],'USB derecho · borde reforzado'),([-0.5,0.85,0.35],'Tapa a presión · sin tornillos')]):
    cam=np.array(cam);cam/=np.linalg.norm(cam)
    right=np.cross([0,0,1],cam);right/=np.linalg.norm(right)
    up=np.cross(cam,right)
    width,height=600,740
    pixels=np.full((height,width,3),[236,233,228],dtype=np.uint8)
    depth=np.full((height,width),-np.inf)
    scale=7.5
    center=np.array([0,11,20])
    for man,color in scene:
        t=mesh(man);world=upright(t);v=world-center
        projected=np.column_stack((v@right*scale+width/2,-v@up*scale+height/2,v@cam))
        for tri,world_tri in zip(projected[t.faces],world[t.faces]):
            n=-np.cross(world_tri[1]-world_tri[0],world_tri[2]-world_tri[0]);n/=max(np.linalg.norm(n),1e-9)
            if n@cam<=0: continue
            x0=max(0,int(np.floor(tri[:,0].min())));x1=min(width-1,int(np.ceil(tri[:,0].max())))
            y0=max(0,int(np.floor(tri[:,1].min())));y1=min(height-1,int(np.ceil(tri[:,1].max())))
            if x1<x0 or y1<y0: continue
            a,b,c=tri
            den=(b[1]-c[1])*(a[0]-c[0])+(c[0]-b[0])*(a[1]-c[1])
            if abs(den)<1e-8: continue
            yy,xx=np.mgrid[y0:y1+1,x0:x1+1];xx=xx+.5;yy=yy+.5
            w0=((b[1]-c[1])*(xx-c[0])+(c[0]-b[0])*(yy-c[1]))/den
            w1=((c[1]-a[1])*(xx-c[0])+(a[0]-c[0])*(yy-c[1]))/den
            w2=1-w0-w1
            z=w0*a[2]+w1*b[2]+w2*c[2]
            d=depth[y0:y1+1,x0:x1+1]
            visible=(w0>=-1e-7)&(w1>=-1e-7)&(w2>=-1e-7)&(z>d)
            shade=.55+.45*max(0,float(n@np.array([-.3,-.5,.812])))
            rgb=np.clip(np.array(to_rgb(color))*shade*255,0,255).astype(np.uint8)
            pixels[y0:y1+1,x0:x1+1][visible]=rgb
            d[visible]=z[visible]
    CANVAS.paste(Image.fromarray(pixels),(20+640*view,120))
    draw.text((320+640*view,110),title,font=font(23),fill='#292929',anchor='mm')
draw.text((960,880),'Dos piezas · tapa con trabas liberables · salidas de luz abiertas',font=font(22),fill='#292929',anchor='mm')
CANVAS.save(ROOT/'vista_previa.png')
print(json.dumps(report,indent=2))
