#!/usr/bin/env python3
"""从参考项目 HanaBi-ui-Genshin 的原始素材，生成 Go 版需要的派生壁纸资源。

这些资源是**本地生成**的，生成规则全部写在下面，可以随时重跑：

    python scripts/assetgen/generate.py --ref D:/workspace/opensource/HanaBi-ui-Genshin

产物（相对 assets/）：

    <id>/bg.jpg      八重神子 / 宵宫：把 .hbc 里的人物合成进背景（原项目的背景图是空场景，
                     人物是运行时单独一层网格画的，Go 版把人物烤进图里）
    <id>/person.png  人物闸门：只有这里的像素才参与呼吸
    <id>/hair.png    发丝/衣摆：R = 基础层（慢、小），G = 细缕层（快、略大）
    <id>/ambient.png 环境漂移：只在大块柔和的区域（天空、光、雾）里动，硬边建筑不动
    shared/ambient.png 全覆盖的漂移遮罩（整幅都是柔光的主题用）

为什么人物要烤进背景图：mygo 的 Painter 没有三角网格绘制，.hbc 是「分层贴图 + 逐帧顶点
位移」的网格（Live2D 类），逐顶点变形没法在片元着色器里做。烤一张中性姿势的合成图，
再用人物遮罩做呼吸/发丝位移，观感与原项目一致，且和甘雨走同一套流程。

依赖：numpy + Pillow。
"""

from __future__ import annotations

import argparse
import io
import os
import struct
import sys

import numpy as np
from PIL import Image, ImageFilter

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
ASSETS = os.path.join(ROOT, "assets")
DEFAULT_REF = r"D:\workspace\opensource\HanaBi-ui-Genshin"

# 遮罩统一 1920×1080（16:9）。着色器按「图片 UV × 遮罩像素尺寸」采样，
# 所以遮罩分辨率和背景分辨率不必相同；边缘靠双线性插值，1/2 分辨率足够软。
MASK_W, MASK_H = 1920, 1080


# ---------------------------------------------------------------- 基础工具


def feather(m: np.ndarray, radius: float) -> np.ndarray:
    """高斯羽化一个 0..1 的遮罩（羽化半径单位是遮罩像素）。"""
    if radius <= 0:
        return m
    img = Image.fromarray((np.clip(m, 0.0, 1.0) * 255).astype(np.uint8))
    img = img.filter(ImageFilter.GaussianBlur(radius))
    return np.asarray(img, dtype=np.float32) / 255.0


def resize_mask(m: np.ndarray, w: int, h: int) -> np.ndarray:
    img = Image.fromarray((np.clip(m, 0.0, 1.0) * 255).astype(np.uint8))
    return np.asarray(img.resize((w, h), Image.LANCZOS), dtype=np.float32) / 255.0


def save_gray(m: np.ndarray, path: str) -> None:
    os.makedirs(os.path.dirname(path), exist_ok=True)
    Image.fromarray((np.clip(m, 0.0, 1.0) * 255).astype(np.uint8), "L").save(path, optimize=True)


def save_rgb(r: np.ndarray, g: np.ndarray, b: np.ndarray, path: str) -> None:
    os.makedirs(os.path.dirname(path), exist_ok=True)
    arr = np.stack([np.clip(r, 0, 1), np.clip(g, 0, 1), np.clip(b, 0, 1)], axis=-1)
    # A 固定 255：mygo 解码时按 alpha 预乘 RGB，A<255 会把遮罩数值乘小。
    rgba = np.concatenate([arr, np.ones_like(arr[..., :1])], axis=-1)
    Image.fromarray((rgba * 255).astype(np.uint8), "RGBA").save(path, optimize=True)


def ellipse(w: int, h: int, cx: float, cy: float, rx: float, ry: float, soft: float = 0.35) -> np.ndarray:
    """归一化坐标（0..1）的柔边椭圆遮罩，soft 是边缘过渡宽度（相对半径）。"""
    y, x = np.mgrid[0:h, 0:w].astype(np.float32)
    nx = (x + 0.5) / w
    ny = (y + 0.5) / h
    d = np.sqrt(((nx - cx) / max(rx, 1e-6)) ** 2 + ((ny - cy) / max(ry, 1e-6)) ** 2)
    return np.clip((1.0 - d) / max(soft, 1e-6), 0.0, 1.0).astype(np.float32)


def detail_energy(img: Image.Image, scale: int = 4) -> np.ndarray:
    """局部对比度（细节强度），0..1。用来把「锐利的人物」和「虚化的背景」分开。"""
    w, h = img.size
    small = img.convert("RGB").resize((max(w // scale, 16), max(h // scale, 16)), Image.LANCZOS)
    a = np.asarray(small, dtype=np.float32) / 255.0
    luma = a @ np.array([0.299, 0.587, 0.114], np.float32)
    blur = np.asarray(
        Image.fromarray((luma * 255).astype(np.uint8)).filter(ImageFilter.GaussianBlur(9)),
        dtype=np.float32,
    ) / 255.0
    contrast = np.abs(luma - blur)
    gx = np.abs(np.diff(luma, axis=1, prepend=luma[:, :1]))
    gy = np.abs(np.diff(luma, axis=0, prepend=luma[:1, :]))
    e = contrast * 2.0 + np.sqrt(gx * gx + gy * gy)
    e = np.asarray(
        Image.fromarray((np.clip(e * 4, 0, 1) * 255).astype(np.uint8)).filter(ImageFilter.GaussianBlur(6)),
        dtype=np.float32,
    ) / 255.0
    return resize_mask(e, w // scale, h // scale), (w // scale, h // scale)


def dilate(m: np.ndarray, r: int) -> np.ndarray:
    """方形结构元的膨胀（可分离实现：先横后竖，各 r 次 ±1 位移取 max）。"""
    out = m
    for axis in (0, 1):
        for k in range(1, max(r, 0) + 1):
            out = np.maximum(out, np.roll(out, k, axis=axis))
            out = np.maximum(out, np.roll(out, -k, axis=axis))
    return out


def softness_mask(img: Image.Image, margin: int = 6) -> np.ndarray:
    """大块柔和区域（天空/光/雾）的遮罩：细节越少权重越高。

    margin 会沿着所有硬边往外扩一圈再扣掉：细线（钟离的菱形光框、珐露珊的
    窗框）本身细节很高、权重本来就接近 0，但紧挨着它的几个像素是"半柔"的，
    漂移起来会看到细线在抖。挖掉一条边距之后，只有真正的大块柔光才动。
    """
    e, _ = detail_energy(img)
    hi = float(np.percentile(e, 95)) or 1.0
    n = np.clip(e / hi, 0.0, 1.0)
    # smoothstep(0.25, 0.8, n) 的反相
    t = np.clip((n - 0.25) / 0.55, 0.0, 1.0)
    soft = 1.0 - (t * t * (3.0 - 2.0 * t))
    if margin > 0:
        edge = (n > 0.30).astype(np.float32)
        near = feather(dilate(edge, margin), 1.5)
        soft = soft * (1.0 - near)
    return soft


def largest_component(m: np.ndarray) -> np.ndarray:
    """只保留最大的连通块（去掉零散的背景碎片）。"""
    h, w = m.shape
    seen = np.zeros((h, w), bool)
    best: list[tuple[int, int]] = []
    for y0 in range(0, h, 6):
        for x0 in range(0, w, 6):
            if m[y0, x0] <= 0 or seen[y0, x0]:
                continue
            stack = [(y0, x0)]
            seen[y0, x0] = True
            comp = []
            while stack:
                y, x = stack.pop()
                comp.append((y, x))
                for dy, dx in ((1, 0), (-1, 0), (0, 1), (0, -1)):
                    ny, nx = y + dy, x + dx
                    if 0 <= ny < h and 0 <= nx < w and m[ny, nx] > 0 and not seen[ny, nx]:
                        seen[ny, nx] = True
                        stack.append((ny, nx))
            if len(comp) > len(best):
                best = comp
    out = np.zeros((h, w), np.float32)
    for y, x in best:
        out[y, x] = 1.0
    return out


def fill_holes(m: np.ndarray) -> np.ndarray:
    """填掉连通块内部的洞（眼睛、高光这些被围住的区域）。"""
    h, w = m.shape
    inv = (m <= 0)
    seen = np.zeros((h, w), bool)
    stack = []
    for x in range(w):
        for y in (0, h - 1):
            if inv[y, x] and not seen[y, x]:
                seen[y, x] = True
                stack.append((y, x))
    for y in range(h):
        for x in (0, w - 1):
            if inv[y, x] and not seen[y, x]:
                seen[y, x] = True
                stack.append((y, x))
    while stack:
        y, x = stack.pop()
        for dy, dx in ((1, 0), (-1, 0), (0, 1), (0, -1)):
            ny, nx = y + dy, x + dx
            if 0 <= ny < h and 0 <= nx < w and inv[ny, nx] and not seen[ny, nx]:
                seen[ny, nx] = True
                stack.append((ny, nx))
    holes = inv & ~seen
    return np.where(holes, 1.0, m).astype(np.float32)


def person_from_detail(img: Image.Image, pct: float, close_radius: float = 2.0,
                       region: tuple[float, float, float, float] | None = None) -> np.ndarray:
    """锐利区域 → 人物遮罩，再填洞、羽化；region 是归一化 (cx,cy,rx,ry) 的粗范围。"""
    e, (w, h) = detail_energy(img)
    thr = float(np.percentile(e, pct))
    m = (e > thr).astype(np.float32)
    m = largest_component(m)
    m = fill_holes(m)
    if region is not None:
        gate = ellipse(w, h, *region, soft=0.6)
        m = m * gate
    # 闭运算：先膨胀后收缩，边缘更连贯
    if close_radius > 0:
        m = feather(m, close_radius)
        m = (m > 0.35).astype(np.float32)
    m = feather(m, 1.4)
    return resize_mask(m, MASK_W, MASK_H)


# ---------------------------------------------------------------- HBC（分层人物网格）


def parse_hbc(path: str) -> dict:
    """解析 HBC0001 容器。结构照 src/gfx/character.cpp。"""
    data = open(path, "rb").read()
    if data[:8] != b"HBC0001\x00":
        raise ValueError(f"{path}: 不是 HBC0001")
    o = 8
    count, period, frames = struct.unpack_from("<IfI", data, o)
    o += 12
    layers = []
    for _ in range(count):
        lid, vcount, icount, png_size, mask_size, kind = struct.unpack_from("<IIIIII", data, o)
        o += 24
        wave = struct.unpack_from("<4f", data, o)
        o += 16
        png = data[o:o + png_size]
        o += png_size
        mask_bytes = data[o:o + mask_size]
        o += mask_size
        verts = np.frombuffer(data[o:o + vcount * 16], dtype="<f4").reshape(vcount, 4).copy()
        o += vcount * 16
        idx = np.frombuffer(data[o:o + icount * 2], dtype="<u2").reshape(-1, 3).copy()
        o += icount * 2
        poses = np.frombuffer(data[o:o + frames * vcount * 4], dtype="<i2").reshape(frames, vcount, 2).copy()
        o += frames * vcount * 4
        layers.append(dict(
            id=lid, kind=kind, wave=wave, verts=verts, idx=idx, poses=poses,
            img=Image.open(io.BytesIO(png)).convert("RGBA"),
            mask=Image.open(io.BytesIO(mask_bytes)).convert("RGBA") if mask_size > 100 else None,
        ))
    if o != len(data):
        raise ValueError(f"{path}: 读了 {o} 字节，文件有 {len(data)} 字节")
    return dict(period=period, frames=frames, layers=layers)


def posed_vertices(layer: dict, frames: int, frame: float) -> np.ndarray:
    a = int(np.floor(frame))
    b = min(a + 1, frames - 1)
    k = frame - a
    pa = layer["poses"][a].astype(np.float32)
    pb = layer["poses"][b].astype(np.float32)
    return layer["verts"][:, :2] + (pa * (1 - k) + pb * k) / 32.0


def raster_layer(layer: dict, frames: int, frame: float, w: int, h: int,
                 with_mask: bool = True):
    """把一个图层光栅化成 (预乘 RGBA 画布, alpha, mask.r)，坐标是 3840×2160 标尺。

    HBC 的顶点坐标固定按 3840×2160 归一化（见 character.cpp 的 VS）。
    """
    pos = posed_vertices(layer, frames, frame)
    uv = layer["verts"][:, 2:4]
    tex_img = np.asarray(layer["img"], dtype=np.float32) / 255.0
    th, tw = tex_img.shape[:2]
    mask_img = None
    if with_mask and layer["mask"] is not None:
        mask_img = np.asarray(layer["mask"], dtype=np.float32) / 255.0
        mh, mw = mask_img.shape[:2]

    rgba = np.zeros((h, w, 4), np.float32)  # 预乘
    alpha = np.zeros((h, w), np.float32)
    maskr = np.zeros((h, w), np.float32)

    for tri in layer["idx"]:
        x = pos[tri, 0]
        y = pos[tri, 1]
        x0 = max(int(np.floor(x.min())), 0)
        x1 = min(int(np.ceil(x.max())) + 1, w)
        y0 = max(int(np.floor(y.min())), 0)
        y1 = min(int(np.ceil(y.max())) + 1, h)
        if x1 <= x0 or y1 <= y0:
            continue
        ax, ay = x[0], y[0]
        bx, by = x[1], y[1]
        cx, cy = x[2], y[2]
        area = (bx - ax) * (cy - ay) - (by - ay) * (cx - ax)
        if abs(area) < 1e-9:
            continue
        px = np.arange(x0, x1, dtype=np.float32) + 0.5
        py = np.arange(y0, y1, dtype=np.float32) + 0.5
        gx, gy = np.meshgrid(px, py)
        w0 = ((bx - gx) * (cy - gy) - (by - gy) * (cx - gx)) / area
        w1 = ((cx - gx) * (ay - gy) - (cy - gy) * (ax - gx)) / area
        w2 = 1.0 - w0 - w1
        inside = (w0 >= -1e-5) & (w1 >= -1e-5) & (w2 >= -1e-5)
        if not inside.any():
            continue
        u = w0 * uv[tri[0], 0] + w1 * uv[tri[1], 0] + w2 * uv[tri[2], 0]
        v = w0 * uv[tri[0], 1] + w1 * uv[tri[1], 1] + w2 * uv[tri[2], 1]
        sx = np.clip(u * (tw - 1), 0, tw - 1.001)
        sy = np.clip(v * (th - 1), 0, th - 1.001)
        ix = sx.astype(np.int32)
        iy = sy.astype(np.int32)
        # 双线性
        fx = (sx - ix)[..., None]
        fy = (sy - iy)[..., None]
        ia = tex_img[iy, ix]
        ib = tex_img[iy, np.minimum(ix + 1, tw - 1)]
        ic = tex_img[np.minimum(iy + 1, th - 1), ix]
        id_ = tex_img[np.minimum(iy + 1, th - 1), np.minimum(ix + 1, tw - 1)]
        texel = ((ia * (1 - fx) + ib * fx) * (1 - fy) + (ic * (1 - fx) + id_ * fx) * fy)
        src_a = texel[..., 3] * inside
        sa3 = src_a[..., None]
        sub = rgba[y0:y1, x0:x1]
        out_a = src_a + sub[..., 3] * (1 - src_a)
        sub[..., :3] = texel[..., :3] * sa3 + sub[..., :3] * (1 - sa3)
        sub[..., 3] = out_a
        alpha[y0:y1, x0:x1] = np.maximum(alpha[y0:y1, x0:x1], src_a)
        if mask_img is not None:
            mx = np.clip(u * (mw - 1), 0, mw - 1).astype(np.int32)
            my = np.clip(v * (mh - 1), 0, mh - 1).astype(np.int32)
            mv = mask_img[my, mx, 0] * src_a
            maskr[y0:y1, x0:x1] = np.maximum(maskr[y0:y1, x0:x1], mv)
    return rgba, alpha, maskr


def bake_character(hbc: dict, frame_ratio: float, w: int, h: int):
    """按图层顺序把人物合成成一张 RGBA，并返回人物 alpha / 发丝遮罩（1920×1080）。

    图层 kind：0 = 普通，1 = 眼睛（视线偏移），2 = 光晕，3 = 裁剪区域
    （后续图层被裁到这个多边形内，直到 id 等于 wave[0] 的图层）。
    """
    frames = hbc["frames"]
    frame = frame_ratio * (frames - 1)
    person = np.zeros((MASK_H, MASK_W), np.float32)
    veil = np.zeros((MASK_H, MASK_W), np.float32)
    best = (0.0, -1)
    sway_fields: list[tuple[int, np.ndarray]] = []

    canvas = np.zeros((h, w, 4), np.float32)  # 预乘
    clip: np.ndarray | None = None
    clip_end = -1

    for i, layer in enumerate(hbc["layers"]):
        rgba, alpha, maskr = raster_layer(layer, frames, frame, w, h, with_mask=True)
        if layer["kind"] == 3:
            clip = alpha > 0.5
            clip_end = int(layer["wave"][0])
            continue
        # 用**累积** alpha 合成，不能用 max：一个图层内部相邻三角形在接缝处会
        # 有半像素重叠，max 会把已经叠加过两次的颜色配上一次 alpha，接缝上出现
        # 一圈发丝描边。另外 raster_layer 返回的 RGB 已经是预乘的，这里不能再乘
        # 一次 alpha——否则半透明边缘被平方，人物轮廓上会浮出一圈暗线（第一批
        # 截图里那些细描边就是这么来的）。
        a = rgba[..., 3]
        clip3 = None
        if clip is not None:
            a = a * clip
            clip3 = clip[..., None]
        sa = a[..., None]
        src = rgba[..., :3]
        if clip3 is not None:
            src = src * clip3
        canvas[..., :3] = src + canvas[..., :3] * (1 - sa)
        canvas[..., 3:4] = sa + canvas[..., 3:4] * (1 - sa)

        small_a = resize_mask(np.clip(a, 0.0, 1.0), MASK_W, MASK_H)
        person = np.maximum(person, small_a)
        if layer["kind"] == 0 and layer["wave"][2] > 0:
            field = resize_mask(maskr, MASK_W, MASK_H) * small_a
            sway_fields.append((i, field))
            total = float(field.sum())
            if total > best[0]:
                best = (total, i)
        if layer["id"] == clip_end:
            clip = None
            clip_end = -1

    if sway_fields:
        for _, field in sway_fields:
            veil = np.maximum(veil, field)
    wave = np.zeros_like(veil)
    if best[1] >= 0:
        for i, field in sway_fields:
            if i == best[1]:
                wave = field
                break
    person = feather(person, 1.0)
    # 羽化大一点：位移幅度 0.003 在 4K 上约 6.5px，遮罩边太硬会在轮廓上"折"出一条
    # 拉伸带；2px 左右的过渡（遮罩分辨率）刚好把形变摊开，人物看着还是实的。
    veil = feather(veil, 2.2)
    wave = feather(wave, 1.8)

    # 反预乘，存成普通 RGBA
    a = canvas[..., 3:4]
    rgb = np.where(a > 1e-6, canvas[..., :3] / np.maximum(a, 1e-6), 0.0)
    char = Image.fromarray((np.concatenate([rgb, a], axis=-1) * 255).round().astype(np.uint8), "RGBA")
    return char, person, veil, wave, best[1]


# ---------------------------------------------------------------- HBW（分层壁纸）


def parse_hbw(path: str) -> dict:
    """解析 HBW0001 容器（钟离）。结构照 src/gfx/layered_wallpaper.cpp。"""
    data = open(path, "rb").read()
    if data[:8] != b"HBW0001\x00":
        raise ValueError(f"{path}: 不是 HBW0001")
    o = 8
    count = struct.unpack_from("<I", data, o)[0]
    o += 4
    noise_size, normal_size = struct.unpack_from("<2I", data, o)
    o += 8
    o += noise_size + normal_size
    layers = []
    for _ in range(count):
        lid, img_size, mask_size = struct.unpack_from("<3I", data, o)
        o += 12
        params = np.frombuffer(data[o:o + 44 * 4], dtype="<f4").copy()
        o += 44 * 4
        img = data[o:o + img_size]
        o += img_size
        mask = data[o:o + mask_size]
        o += mask_size
        layers.append(dict(
            id=lid, params=params,
            img=Image.open(io.BytesIO(img)).convert("RGBA"),
            mask=Image.open(io.BytesIO(mask)).convert("RGBA"),
        ))
    if o != len(data):
        raise ValueError(f"{path}: 读了 {o} 字节，文件有 {len(data)} 字节")
    return dict(layers=layers, size=(2559, 1428))


def place_mask(src: np.ndarray, rect: tuple[float, float, float, float], canvas_hw: tuple[int, int],
               src_size: tuple[int, int]) -> np.ndarray:
    """把源图空间里的一块（rect，单位是源图像素）贴到归一化的遮罩画布上。"""
    cw, ch = canvas_hw
    sw, sh = src_size
    x, y, w, h = rect
    layer = Image.fromarray((np.clip(src, 0, 1) * 255).astype(np.uint8), "L")
    layer = layer.resize((max(int(round(w / sw * cw)), 1), max(int(round(h / sh * ch)), 1)), Image.LANCZOS)
    canvas = Image.new("L", (cw, ch), 0)
    canvas.paste(layer, (int(round(x / sw * cw)), int(round(y / sh * ch))))
    return np.asarray(canvas, dtype=np.float32) / 255.0


# ---------------------------------------------------------------- 各主题的生成规则


def gen_ganyu(ref: str, out: str) -> None:
    """甘雨：沿用原项目的三张遮罩，打包成 person + hair（R=基础层，G=细缕层）。"""
    person = Image.open(os.path.join(ref, "ganyu-shake-opacity.png")).convert("L")
    veil = Image.open(os.path.join(ref, "ganyu-flow-mask.png")).convert("L")
    wave = Image.open(os.path.join(ref, "ganyu-wave-a.png")).convert("L")
    save_gray(resize_mask(np.asarray(person, np.float32) / 255.0, MASK_W, MASK_H),
              os.path.join(out, "ganyu", "person.png"))
    save_rgb(resize_mask(np.asarray(veil, np.float32) / 255.0, MASK_W, MASK_H),
             resize_mask(np.asarray(wave, np.float32) / 255.0, MASK_W, MASK_H),
             np.zeros((MASK_H, MASK_W), np.float32),
             os.path.join(out, "ganyu", "hair.png"))


def gen_character_theme(ref: str, out: str, tid: str, bg_name: str, hbc_name: str,
                        frame_ratio: float, hair_extra=None, ambient_softness: bool = True) -> None:
    """八重神子 / 宵宫：把 .hbc 人物烤进背景，并生成人物/发丝遮罩。"""
    bg = Image.open(os.path.join(ref, bg_name)).convert("RGB")
    w, h = bg.size
    hbc = parse_hbc(os.path.join(ref, hbc_name))
    char, person, veil, wave = bake_character(hbc, frame_ratio, w, h)[:4]
    if hair_extra is not None:
        veil, wave = hair_extra(person, veil, wave)
    comp = Image.alpha_composite(bg.convert("RGBA"), char)
    os.makedirs(os.path.join(out, tid), exist_ok=True)
    comp.convert("RGB").save(os.path.join(out, tid, "bg.jpg"), quality=95, subsampling=0, optimize=True)
    save_gray(person, os.path.join(out, tid, "person.png"))
    save_rgb(veil, wave, np.zeros_like(veil), os.path.join(out, tid, "hair.png"))
    if ambient_softness:
        soft = resize_mask(softness_mask(comp), MASK_W, MASK_H)
        save_gray(feather(soft, 2.0), os.path.join(out, tid, "ambient.png"))
    total = len(person[person > 0.02])
    print(f"  {tid}: 人物 {total / person.size * 100:.1f}%，发丝基础层 {veil.mean() * 100:.1f}%，"
          f"细缕层 {wave.mean() * 100:.1f}%")


def gen_yoimiya(ref: str, out: str) -> None:
    """宵宫：人物烤进背景；.hbc 的每层遮罩是空的（1×1），发丝区按头部椭圆取。

    她的脸也在头部椭圆里，必须挖掉：发丝位移一旦扫过脸，五官会被拉变形
    （第一批截图里她的脸就是被拉歪的，第一版还没意识到是发丝遮罩盖到了脸）。
    """

    def hair(person, veil, wave):
        head = ellipse(MASK_W, MASK_H, 0.505, 0.225, 0.115, 0.150, soft=0.55)
        tail = ellipse(MASK_W, MASK_H, 0.435, 0.300, 0.065, 0.105, soft=0.60)
        # 脸在 y≈26%~36%（眼睛 26~29%、嘴 34%）。椭圆要盖住整张脸（含眼角），
        # 发丝位移一旦扫过脸，五官会被拉变形。
        face = ellipse(MASK_W, MASK_H, 0.495, 0.305, 0.075, 0.090, soft=0.45)
        veil = person * np.maximum(head, 0.0) * (1.0 - face)
        wave = person * tail * (1.0 - face)
        return veil, wave

    gen_character_theme(ref, out, "yoimiya", "yoimiya-bg.jpg", "yoimiya-character.hbc", 0.28, hair_extra=hair)


def gen_zhongli(ref: str, out: str) -> None:
    """钟离：壁纸是 .hbw 分层容器，人物是第 7 层，金色漩涡是第 8 层。

    背景仍用现有 bg.jpg（作者已从 hbw 合成过），这里只补人物/发丝/环境遮罩。
    实测第 7 层 (id=85) 的 alpha 与现有 bg.jpg 里的人物像素级对齐。
    """
    hbw = parse_hbw(os.path.join(ref, "zhongli-wallpaper.hbw"))
    sw, sh = hbw["size"]
    char_layer = next(L for L in hbw["layers"] if L["id"] == 85)
    orb_layer = next(L for L in hbw["layers"] if L["id"] == 35)
    rect = tuple(float(v) for v in char_layer["params"].reshape(-1, 4)[0])
    orb_rect = tuple(float(v) for v in orb_layer["params"].reshape(-1, 4)[0])

    alpha = np.asarray(char_layer["img"], dtype=np.float32)[..., 3] / 255.0
    maskr = np.asarray(char_layer["mask"], dtype=np.float32)[..., 0] / 255.0
    person = place_mask(alpha, rect, (MASK_W, MASK_H), (sw, sh))
    hair = place_mask(maskr * alpha, rect, (MASK_W, MASK_H), (sw, sh))
    orb = np.asarray(orb_layer["img"], dtype=np.float32)[..., 3] / 255.0
    glow = place_mask(orb, orb_rect, (MASK_W, MASK_H), (sw, sh))

    bg = Image.open(os.path.join(out, "zhongli", "bg.jpg")).convert("RGB")
    soft = resize_mask(softness_mask(bg), MASK_W, MASK_H)
    ambient = np.maximum(feather(soft, 2.0) * 0.85, feather(glow, 1.5))

    save_gray(feather(person, 1.0), os.path.join(out, "zhongli", "person.png"))
    save_rgb(feather(hair, 1.4), np.zeros_like(hair), np.zeros_like(hair), os.path.join(out, "zhongli", "hair.png"))
    save_gray(ambient, os.path.join(out, "zhongli", "ambient.png"))
    print(f"  zhongli: 人物 {person.mean() * 100:.1f}%，发丝 {hair.mean() * 100:.1f}%，"
          f"环境（柔光+金涡）{ambient.mean() * 100:.1f}%")


def gen_painted_theme(ref: str, out: str, tid: str, bg_name: str, person_pct: float,
                      hair_src: str | None, hair_ellipses, ambient_softness: bool = True,
                      region=None, person_extra=None, hair_exclude=None) -> None:
    """纳西妲 / 芙宁娜 / 珐露珊：人物已经画在背景里，遮罩从原项目的动态图和细节图推。"""
    src = os.path.join(ref, bg_name)
    img = Image.open(src).convert("RGB")
    person = person_from_detail(img, person_pct, region=region)
    # 细节图抓不住的部位（大面积纯色裙子之类）补几个椭圆：权重压到 0.9，
    # 边缘本来就羽化，位移在边界上平滑收敛。
    for e in person_extra or []:
        cx, cy, rx, ry, w = e
        person = np.maximum(person, ellipse(MASK_W, MASK_H, cx, cy, rx, ry, soft=0.55) * w)
    person = feather(np.clip(person, 0.0, 1.0), 1.0)
    if hair_src:
        h = np.asarray(Image.open(os.path.join(ref, hair_src)).convert("L"), np.float32) / 255.0
        veil = feather(resize_mask(h, MASK_W, MASK_H), 1.2)
        wave = np.zeros_like(veil)
    else:
        veil = np.zeros((MASK_H, MASK_W), np.float32)
        wave = np.zeros_like(veil)
    for e in hair_ellipses:
        m = ellipse(MASK_W, MASK_H, *e, soft=0.5)
        if e == hair_ellipses[-1] and hair_ellipses:
            wave = np.maximum(wave, m * person)
        else:
            veil = np.maximum(veil, m * person)
    # 挖掉脸：发丝在脸周围飘，但脸本身不该跟着位移，否则五官会和脖子错位。
    for e in hair_exclude or []:
        hole = ellipse(MASK_W, MASK_H, *e, soft=0.45)
        veil = veil * (1.0 - hole)
        wave = wave * (1.0 - hole)
    save_gray(person, os.path.join(out, tid, "person.png"))
    save_rgb(veil, wave, np.zeros_like(veil), os.path.join(out, tid, "hair.png"))
    if ambient_softness:
        save_gray(feather(resize_mask(softness_mask(img), MASK_W, MASK_H), 2.0),
                  os.path.join(out, tid, "ambient.png"))
    print(f"  {tid}: 人物 {person.mean() * 100:.1f}%，发丝 {veil.mean() * 100:.1f}%，"
          f"细缕 {wave.mean() * 100:.1f}%")


def gen_nahida(ref: str, out: str) -> None:
    # nahida-wave.png 是原项目用来做横向波动的遮罩，形状正好是头发/叶片。
    gen_painted_theme(ref, out, "nahida", "nahida-bg.jpg", 55.0, "nahida-wave.png", [])


def gen_furina(ref: str, out: str) -> None:
    # furina-sway.png 就是她的发梢形状（原项目拿它做水波摆动）。
    gen_painted_theme(ref, out, "furina", "furina-bg.jpg", 60.0, "furina-sway.png", [])


def gen_faruzan(ref: str, out: str) -> None:
    # 原项目只有一张流向图（无遮罩），发丝区域按头部与马尾两个椭圆取。
    #
    # 人物遮罩从「细节强度」推出来时，她的**脸、脖子和马尾**都会被漏掉：
    # 脸和脖子是光滑皮肤（细节少），马尾是大块深色头发（对比低）。漏掉的后果是
    # 头在呼吸、脖子和身体不动，下巴和脖子之间像被裁了一刀——这就是"脖子动得
    # 很奇怪"的根因。所以这里按实测坐标补三块：
    #   - 脸（x 0.375~0.465 / y 0.22~0.34）：补进人物遮罩，头和脸一起起伏；
    #   - 脖子（x 0.41~0.47 / y 0.30~0.39）：把头和身体连起来，避免断开；
    #   - 马尾（x 0.315~0.405 / y 0.26~0.48）：让发丝遮罩有东西可摆。
    # 白裙子几乎和亮背景同色，再补一个身体椭圆，否则呼吸只作用在肩膀以上。
    gen_painted_theme(
        ref, out, "faruzan", "faruzan-bg.jpg", 62.0, None,
        [(0.435, 0.215, 0.085, 0.105), (0.360, 0.370, 0.062, 0.130)],
        region=(0.46, 0.52, 0.30, 0.52),
        person_extra=[
            (0.462, 0.700, 0.150, 0.300, 0.90),  # 身体/裙子
            (0.420, 0.280, 0.045, 0.060, 0.95),  # 脸
            (0.440, 0.345, 0.030, 0.048, 0.95),  # 脖子
            (0.360, 0.370, 0.045, 0.110, 0.90),  # 马尾
        ],
        hair_exclude=[(0.443, 0.260, 0.040, 0.045)],
    )


def gen_shared(out: str) -> None:
    """保留的挂载点：将来若出现「整幅画面都要漂移」的主题，可以在这里生成全白遮罩。

    目前七个主题都有各自的 ambient.png（甘雨不配环境层），所以这里是空实现。
    """
    return


GENERATORS = {
    "ganyu": gen_ganyu,
    "yae": lambda ref, out: gen_character_theme(
        ref, out, "yae", "yae-bg-4k.jpg", "yae-character.hbc", 0.33),
    "yoimiya": gen_yoimiya,
    "nahida": gen_nahida,
    "furina": gen_furina,
    "faruzan": gen_faruzan,
    "zhongli": gen_zhongli,
}


def main() -> int:
    ap = argparse.ArgumentParser(description="生成 HanaBi Go 版的派生壁纸资源")
    ap.add_argument("--ref", default=os.environ.get("HANABI_REF", DEFAULT_REF),
                    help="参考项目 HanaBi-ui-Genshin 的根目录")
    ap.add_argument("--out", default=ASSETS, help="输出目录（默认项目的 assets/）")
    ap.add_argument("--only", default="", help="只生成这些主题，逗号分隔")
    args = ap.parse_args()

    tex = os.path.join(args.ref, "assets", "tex")
    if not os.path.isdir(tex):
        print(f"找不到参考素材目录：{tex}", file=sys.stderr)
        return 1
    only = [x.strip() for x in args.only.split(",") if x.strip()]
    gen_shared(args.out)
    for tid, fn in GENERATORS.items():
        if only and tid not in only:
            continue
        print(f"[{tid}]")
        fn(tex, args.out)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
