# 花火 HanaBi（bestuav-hanabi）

以角色主题为核心的桌面 UI，用 Go + [mygo](https://github.com/egoist/mygo) 实现。
视觉与交互参考开源项目 [HanaBi-ui-Genshin](https://github.com/xiaoyeuzixi/HanaBi-ui-Genshin)（C++/ImGui 原生实现），本项目是它的 Go 重写版本，**只借鉴设计，不包含其资产与代码**。

## 当前阶段：多角色主题 + 开场动画 + 动态壁纸 + 毛玻璃

已完成：

- **多角色主题**：一套「角色包」结构——令牌（11 色 + 圆角）、元素、标识画法（Motif）、壁纸染色强度
  - 主题库点卡片即切换：侧栏、卡片、文字、玻璃 tint、壁纸氛围**同时变**
  - 三套完整主题：甘雨（冰·冰晶标识）、宵宫（火·焰火标识）、雷电将军（雷·雷光标识）
  - 其余四个角色复用冰晶画法，等有素材再补各自 Motif
  - **壁纸染色在着色器里做**（按元素色重映射亮度），不是叠半透明色层——后者会把暗部提亮、糊掉细节
  - 新增角色只改 `internal/theme/theme.go` 一条数据，页面与绘制层不动
- **开场动画（6.20 秒）**：时间轴与粒子参数**照搬原项目**（同一套 LCG 随机数种子，粒子分布逐帧对应）
  - 时间轴：0.20s 上升 → 1.45s 转向 → 3.65s 成形 → 4.65s 停靠 → 6.20s 结束
  - 32 片雪（斜向短划）+ 20 盏孔明灯（暖光晕、边缘淡入淡出），随 6.2 秒向上飘
  - 标识按当前角色的 Motif 画：冰晶（三条 V 形分叉 + 菱形臂尖，**每条臂错开生长**）/ 焰火（放射火舌 + 末端火星，成形时炸开）/ 雷光（折线闪电）
  - 结束时标识收缩并停靠到侧栏头像位置，开场层淡出、主界面「透出来」（不切视图）
  - 点画面任意处可跳过（0.9 秒内快进走完）
- **动态壁纸**：`internal/wallpaper/wallpaper.kage`（kage 着色器）一次绘制完成
  「Cover 铺满 + 随时间漂移的 UV 扰动 + 主题染色」，位移按图片空间计（窗口缩放观感一致），
  近景权重更大形成轻微景深；`Play` 参数可关闭动效。
- **12fps 后台时钟**驱动壁纸（`Painter.AnimationFrameRate`，只重绘那一层、不重建元素树）
- **毛玻璃**：卡片用 `glass.Regular`（Clarity 0.45）、侧栏与顶/底栏用 `glass.Clear`，
  壁纸从玻璃下透出；悬停与选中都用 `Tint`（玻璃的底是壁纸，换底色会破坏玻璃感）
- 顶部渐变压深带（170 DIP 渐隐到透明），让大标题压在花哨壁纸上仍清晰
- 窗口 + 三栏布局 + 整宽底部状态栏

未完成（见路线图）：

- 每个角色各自的背景素材（现在共用甘雨壁纸，靠染色区分）
- 标识的 3D 质感（原版是 5122 三角形 + 自定义顶点着色，这里是矢量近似）
- 卡片缩略图与元素筛选
- 设置页（含「动画开关」，`wallpaper.Draw` 的 `play` 与 `intro` 已预留）
- 打包与自动更新

## 预览

| 主题库 · 甘雨（冰） | 主题库 · 宵宫（火） | 主题库 · 雷电将军（雷） |
|---|---|---|
| ![甘雨](docs/screenshots/theme-ganyu.png) | ![宵宫](docs/screenshots/theme-yoimiya.png) | ![雷电将军](docs/screenshots/theme-raiden.png) |

| 开场 · 抽枝生长（0.9s） | 开场 · 转向成形（2.5s） | 开场 · 完全成形（3.75s） |
|---|---|---|
| ![rise](docs/screenshots/intro-1-rise.png) | ![turn](docs/screenshots/intro-2-turn.png) | ![formation](docs/screenshots/intro-3-formation.png) |

## 已知的坑（写下来免得重犯）

1. **`Painter.Line` 只能画水平或垂直线**。它的文档写明 "straight horizontal or vertical line"，实现是取两点包围盒再用 width 覆盖某一维——**斜线会被压成一个矩形**。冰晶六条斜臂全变成竖线、雪花变成小方块，就是这个原因。任意角度线段要用 `Transform` + `Fill`（见 `intro.go` 的 `seg`）。
2. **`image.RGBA` 是预乘 alpha**。往 `SetRGBA` 里写 `(255,255,255,0)` 这种 RGB>A 的非法值，渲染器会把边缘也当成不透明——光晕变成实心圆盘。必须写 `(v,v,v,v)`。
3. **不能用多层同心圆模拟径向渐变**：每层都是硬边，叠出来是一圈圈台阶，比单层圆更难看。正确做法是程序化生成一张渐变贴图（`ui.NewBitmap`）再用 `ImageRegion` 拉伸 + `Tint`（见 `glow`）。
4. **壁纸染色不能靠叠半透明色层**：那样只会把暗部提亮、整体发灰，蓝紫色像素乘上低亮度几乎不变，换主题时看不出区别。要在着色器里**按元素色重映射亮度**（`luma / tintLuma`），明暗层次才保留。
5. **`Material` 会替换 `Background`**：`glass.Glass` 是材质，同时写 `Background` 是死代码；悬停想加色只能重新 `Material(Tint)`。
6. **离屏截图的画布默认是白色**：开场层底色带透明度，不先铺不透明底会得到"白茫茫"的假象（真实运行时下面压着主界面）。
7. **根容器必须 `Fill()` + `AlignItems(ui.Stretch)`**：只写 `Grow(1)` 时它只占父容器主轴（宽度），子列会按内容高度垂直居中。
8. **资源根不能用「找 assets 目录」来定位**：包目录 `internal/assets` 与资源目录同名，向上查找会误命中。现在优先用 `mygo.App.Path(mygo.PathResources)`，再回退 `go.mod` 锚点。
9. **图片解码必须缓存**：壁纸是 1200 万像素的 JPG，每帧解码会直接把帧率打到个位数（见 `internal/assets`）。
10. **CPU 渲染器跑 kage 很慢**：离屏截图走 CPU 解释器，1280×800 一帧约 0.3s；真机走 GPU 才够 12fps 或 60fps。
11. **资源不随仓库分发**：`assets/` 在 `.gitignore` 里，打包前必须在本地存在（`mygo.json` 的 `resources: ["assets"]` 会把它复制到 exe 旁）。

## 运行

```bash
# 开发（热重载）
go tool mygo dev

# 直接跑
go run .

# 编译校验
go build ./...
```

## 目录

```text
main.go                 启动：窗口 + Content
mygo.json               打包配置（name/slug/version；resources 把 assets/ 复制到 exe 旁）
internal/theme/         角色包：令牌 + 元素 + 标识画法 + 壁纸染色强度
internal/assets/        资源定位（resources 目录 / go.mod 锚点）+ 解码缓存
internal/wallpaper/     wallpaper.kage + Draw：动态壁纸着色器
internal/intro/         开场的时间轴、粒子与标识姿态（纯逻辑，可单测）
internal/ui/icons/      内联 SVG 图标（currentColor 跟随主题）
internal/view/          界面根：App 状态、开场渲染、侧栏、标题栏、页面
docs/screenshots/       界面截图（离屏渲染产出）
```

## 测试

```bash
go test ./...          # 时间轴/粒子/姿态 + 开场渲染冒烟
```

`internal/intro` 的测试钉住时间轴（6.20s 结束、dt 夹到 0.20）、姿态起止点（中央偏下 → 侧栏停靠点半径 16）、
粒子确定性（同一尺寸两次初始化完全一致）、跳过耗时（≤0.9s）以及淡出/交接的单调性。
`internal/wallpaper` 钉住"时间推进会改变画面、`play=0` 时完全静止"。

## 路线图

1. **角色素材**：每个角色自己的壁纸（现在共用甘雨那张，靠染色区分）；有了素材后 `WallTint` 可以降到很低
2. **主题库**：卡片缩略图、按元素筛选
3. **设置**：字号、启动主题、动画开关（`wallpaper.Draw` 的 `play`、开场是否播放）、切换时是否重播开场
4. **打包与更新**：沿用 mygo 的 `build`/`-upload` 与内置 Updater

## 说明

- 角色名称与相关视觉元素归原版权方所有，本项目仅作技术学习与本地演示，不用于商业分发。
- 资产（壁纸、序列帧）不进入版本库，放在 `assets/`（已 gitignore），由本地脚本从原项目导入。
