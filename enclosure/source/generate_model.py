"""Modelo de prueba, mm. pip install manifold3d numpy trimesh matplotlib pillow.
Plano: https://docs.waveshare.com/ESP32-C6-LCD-1.47
Cotas de botones, USB y difusores incluyen margen y requieren prueba fisica.
"""
from pathlib import Path
import json, math, zipfile
import numpy as np
import manifold3d as m
import trimesh
import os
os.environ.setdefault('MPLCONFIGDIR','/private/tmp/claude-esp32/matplotlib-cache')
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt
from mpl_toolkits.mplot3d.art3d import Poly3DCollection

OUT = Path(__file__).resolve().parent
PCB_W, PCB_H, PCB_D = 38.0, 20.0, 9.0 # measurements confirmed by user: includes acrylic and metal standoffs
CW, CH = PCB_W+1.2, PCB_H+1.2 # 0.6 mm assembly clearance per side
W,H = 48.,34.
D = 3.2+PCB_D+0.3 # rear lid stops module with 0.3 mm depth clearance
FRONT = 3.2 # deeper bezel seat, recesses module 1.2 mm more than v2
LCD_W,LCD_H,LCD_X = 32.95,17.99,0.965
USB_SIDE=-1 # right when looking at LCD from front, after standing transform
USB_WIDTH=15.0
USB_HEIGHT=8.8
USB_Z=FRONT+6.0
USB_RIM_BACK=USB_Z+USB_HEIGHT/2+2.4
LIGHT_Z,LIGHT_H = 5.2,1.6
TILT_DEG = 25.0
FOOT_LEVEL = 28.0
SUPPORT_DEPTH = 36.0
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
shape += box(-34,-30.5,-9,9,2.5,USB_RIM_BACK)
for x in [-18,-6,6,18]:
    shape += box(x-3,x+3,-FOOT_LEVEL,-14.5,0,D)
# Rear foot joins all legs at their base; assembled support depth 28 mm.
# Rear base follows the sloped contact plane while retaining 3 mm thickness.
shape += box(-24,24,-FOOT_LEVEL,-FOOT_LEVEL+3,12,SUPPORT_DEPTH).transform(SHEAR)
# Remove material below the tabletop contact plane from all four legs.
shape -= box(-100,100,-100,-FOOT_LEVEL,-10,100).transform(SHEAR)
shape -= box(-CW/2,CW/2,-CH/2,CH/2,FRONT,D+1)
shape -= box(LCD_X-LCD_W/2,LCD_X+LCD_W/2,-LCD_H/2,LCD_H/2,-1,FRONT+0.1)
# BOOT/RST on long edges, near USB end. Broad tool/fingertip access.
BUTTON_X = -(PCB_W/2-7.04)
for s in [-1,1]:
    a,b = sorted([s*(CH/2-0.1),s*(H/2+1)])
    shape -= box(BUTTON_X-4.5,BUTTON_X+4.5,a,b,7.0,D+1)
# USB-C at right. Fits plug envelope provisionally; SD accessible with lid off.
USB_WIDTH=15.0
USB_HEIGHT=8.8 # 6 mm cable thickness plus clearance and position margin
USB_Z=FRONT+6.0
shape -= box(-35,-CW/2+0.1,-USB_WIDTH/2,USB_WIDTH/2,
             USB_Z-USB_HEIGHT/2,USB_Z+USB_HEIGHT/2)
# Light channels through the walls, backed by removable translucent inserts.
for s in [-1,1]:
    a,b=sorted([s*(CH/2-0.1),s*(H/2+1)])
    shape -= box(-14,13,a,b,LIGHT_Z,LIGHT_Z+LIGHT_H)
    a,b=sorted([s*(CW/2-0.1),s*33])
    shape -= box(a,b,-4.5,4.5,LIGHT_Z,LIGHT_Z+LIGHT_H)
# Robust sliding guide rails. No flexing hooks or screws.
# Track gap 2.8 mm; lid thickness 2.4 mm. Side clearance 0.3 mm each.
T=2.4
TRACK_Z=2.8
RAIL_T=2.0
for side in [-1,1]:
    lo,hi=sorted([side*22,side*26])
    shape += box(lo,hi,-17,17,D-1,D)
    lo,hi=sorted([side*24,side*26])
    shape += box(lo,hi,-17,17,D-0.1,D+TRACK_Z+RAIL_T)
    lo,hi=sorted([side*22,side*26])
    shape += box(lo,hi,-17,17,D+TRACK_Z,D+TRACK_Z+RAIL_T)
# Bottom stop holds cover in service; entry at top remains open.
shape += box(-26,26,-19,-16.7,D-1,D+TRACK_Z+RAIL_T)
# Keep USB path clear through the root of the left guide as well.
shape -= box(-35,-CW/2+0.1,-USB_WIDTH/2,USB_WIDTH/2,USB_Z-USB_HEIGHT/2,USB_Z+USB_HEIGHT/2)
# Flat cover with broad integrated pull tab, printed exterior face down.
lid=box(-23.7,23.7,-16.4,16.4,0,T)
lid += box(-6,6,16.3,20,0,T)
for x in [-6,0,6]: lid -= box(x-1,x+1,-6,6,-0.1,T+0.1)

def mesh(man):
    data=man.to_mesh()
    return trimesh.Trimesh(vertices=np.array(data.vert_properties)[:,:3],faces=np.array(data.tri_verts),process=True)
# V6 is the exact opposite USB side from the printed V5 reference photo.
parts={'housing':shape,'rear-cover':lid}
report={'tilt_degrees':TILT_DEG}
for name,man in parts.items():
    t=mesh(man)
    assert t.is_watertight and t.is_volume, name
    assert len(man.decompose())==1, name
    t.apply_translation([0,0,-t.bounds[0,2]])
    t.export(OUT/f'{name}.stl')
    report[name]={'dimensions_mm':t.extents.round(3).tolist(),'triangles':len(t.faces),'volume_mm3':round(float(t.volume),2),'watertight':bool(t.is_watertight),'solids':len(man.decompose())}

# Installed meshes for preview. Assembly preview is not a print file.
cover=lid.translate((0,0,D))
assert (shape ^ cover).volume()<0.001
board_envelope=box(-PCB_W/2,PCB_W/2,-PCB_H/2,PCB_H/2,FRONT+0.01,FRONT+PCB_D)
assert (shape ^ board_envelope).volume()<0.001
assert (cover ^ board_envelope).volume()<0.001
module_with_buttons=board_envelope
for sign in [-1,1]:
    lo,hi=sorted([sign*(PCB_H/2-0.1),sign*(PCB_H/2+1.0)])
    module_with_buttons += box(BUTTON_X-2.5,BUTTON_X+2.5,lo,hi,FRONT+4.4,FRONT+PCB_D)
for distance in [0,0.5,1,2,4,8,12]:
    assert (shape ^ module_with_buttons.translate((0,0,distance))).volume()<0.001
assert (cover ^ module_with_buttons).volume()<0.001
# Cover is inserted downward from above. Validate full insertion/extraction.
withdraw=[]
for step in [0,0.5,1,2,4,8,16,24,36,42]:
    collision=(shape ^ cover.translate((0,step,0))).volume()
    assert collision<0.001,(step,collision)
    withdraw.append(round(collision,6))
# Rail undercuts retain cover against backward pulling without fragile clips.
assert (shape ^ cover.translate((0,0,0.6))).volume()>0.05
cable_reference=box(-36,-PCB_W/2,-6,6,USB_Z-3,USB_Z+3)
assert (shape ^ cable_reference).volume()<0.001
assert (cover ^ cable_reference).volume()<0.001
report['assembly']={'housing_cover_intersection_mm3':round((shape^cover).volume(),5),
                   'sliding_path_intersections_mm3':withdraw,
                   'guide_thickness_mm':RAIL_T,'cover_depth_clearance_mm':TRACK_Z-T,
                   'usb_opposite_side_from_printed_v5':USB_SIDE<0,
                   'cavity_mm':[CW,CH,D-FRONT],
                   'measured_module_mm':[PCB_W,PCB_H,PCB_D],
                   'module_depth_clearance_mm':0.3,
                   'usb_opening_mm':[USB_WIDTH,USB_HEIGHT]}
# Check the real assembled contact plane, including every foot and rear base.
vertices=mesh(shape).vertices
angle=math.radians(TILT_DEG)
heights=math.cos(angle)*(vertices[:,1]+FOOT_LEVEL)-math.sin(angle)*vertices[:,2]
assert heights.min()>-1e-5
contact=vertices[np.abs(heights)<1e-5]
assert len(contact)>=12
report['support']={'minimum_height_mm':round(float(heights.min()),6),
                 'coplanar_vertices':len(contact),'support_depth_mm':round(SUPPORT_DEPTH/math.cos(angle),3)}
face_up_normal=np.array([0,math.cos(angle),math.sin(angle)])
assert face_up_normal[2]>0
floor_faces=mesh(shape).faces[np.all(np.abs(heights[mesh(shape).faces])<1e-5,axis=1)]
assert len(floor_faces)>=8
report['support']['screen_upward_tilt_degrees']=round(math.degrees(math.asin(face_up_normal[2])),3)
report['support']['coplanar_contact_triangles']=len(floor_faces)
(OUT/'validation.json').write_text(json.dumps(report,indent=2))

# Render geometry directly; opaque orange shell and milk-white diffuser pieces.
def upright(t):
    verts=t.vertices.copy()
    # The table plane becomes Z=0; screen normal points upward by TILT_DEG.
    c=math.cos(math.radians(TILT_DEG));s=math.sin(math.radians(TILT_DEG))
    h=verts[:,1]+FOOT_LEVEL
    return np.column_stack((verts[:,0],-c*verts[:,2]-s*h,c*h-s*verts[:,2]))

# Orthographic z-buffer renderer avoids painter-order artifacts on CAD cavities.
from PIL import Image, ImageDraw, ImageFont
from matplotlib.colors import to_rgb
CANVAS=Image.new('RGB',(1920,920),'#ece9e4')
def font(size):
    return ImageFont.truetype('/System/Library/Fonts/Supplemental/Arial.ttf',size)
draw=ImageDraw.Draw(CANVAS)
draw.text((960,40),'ESP32-C6 LCD 1.47 · v7 · Desktop viewing angle · 25°',font=font(32),fill='#292929',anchor='mm')
scene=[(shape,'#ed6b32'),(cover,'#d95023')]+[(box(LCD_X-LCD_W/2,LCD_X+LCD_W/2,-LCD_H/2,LCD_H/2,FRONT+0.12,FRONT+0.4),'#15232c')]
for view,(cam,title) in enumerate([([-0.5,0.85,0.35],'Front view · upward-facing display'),([-1,0,0.12],'Side view · 25° backward tilt'),([0.5,-0.85,0.35],'Sliding cover · reinforced guides')]):
    cam=np.array(cam);cam/=np.linalg.norm(cam)
    right=np.cross([0,0,1],cam);right/=np.linalg.norm(right)
    up=np.cross(cam,right)
    width,height=600,740
    pixels=np.full((height,width,3),[236,233,228],dtype=np.uint8)
    depth=np.full((height,width),-np.inf)
    scale=7.5
    center=np.array([0,-11,20])
    for man,color in scene:
        if False: continue
        t=mesh(man);world=upright(t);v=world-center
        projected=np.column_stack((v@right*scale+width/2,-v@up*scale+height/2,v@cam))
        for tri,world_tri in zip(projected[t.faces],world[t.faces]):
            n=np.cross(world_tri[1]-world_tri[0],world_tri[2]-world_tri[0]);n/=max(np.linalg.norm(n),1e-9)
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
draw.text((960,880),'25° backward tilt · Flat tabletop support · Sliding rear cover',font=font(22),fill='#292929',anchor='mm')
CANVAS.save(OUT/'preview.png')
print(json.dumps(report,indent=2))
