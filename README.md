# 花火 HanaBi（bestuav-hanabi）

以角色主题为核心的桌面 UI，用 Go + [mygo](https://github.com/egoist/mygo) 实现。
视觉与交互参考开源项目 [HanaBi-ui-Genshin](https://github.com/xiaoyeuzixi/HanaBi-ui-Genshin)（C++/ImGui 原生实现）。本项目是它的 Go 重写版本：**代码全部重写，只借鉴设计与交互**；壁纸素材由本地从原项目导入，不随仓库分发。

## 当前阶段：多角色主题 + 开场动画 + 动态壁纸 + 设置、转场与自动更新

已完成：

- **无系统标题栏**：`TitleBarHidden` + `HideWindowButtons`，标题栏与最小化/最大化/关闭全部自绘，
  整条可控拖动（`DragWindow`，双击最大化）
- **侧栏三段式**：品牌（元素标识 + HanaBi）/ 分组标签 + 菜单（图标 + 文字 + 选中项右侧的小标识；
  带纹样的主题用专属花饰）/
  底部当前角色大标识 + 名字。品牌图标的位置正是开场标识的停靠点，交接时不跳变
- **主题页**：大标题 + `APPEARANCE` 小标 + 副标题，右上分段控件，下方元素筛选条
  （全部 / 冰 / 水 / 雷 / 火 / 草 / 风 / 岩 + 角色数）；**只有卡片滚动**，标题与筛选条固定；
  卡片为**纯色底**：七个角色各一套纹样（金鱼+烟花、樱花+御守、清心花+中国结、叶枝、
  礼帽+水波、棱晶+古书、岩印+回纹），元素徽记统一放大到卡片内 (36,31)；
  **列数随窗口宽度自适应**（2~4 列，卡片行内撑满、末行保持同宽，flex-wrap 效果）；
  英文名用衬线体 + 「中文名·属性」（属性用元素色），当前主题带元素色高亮描边；
  状态栏只属于内容区（不通到侧栏），侧栏底部居中
- **多角色主题**：一套「角色包」结构——令牌（11 色 + 圆角）、元素、标识画法（Motif）、
  窗口按钮造型（Chrome）、纹样（Ornament）、命之座、壁纸染色强度
  - 主题库点卡片即切换：侧栏、卡片、文字、卡片 tint、壁纸氛围一起变，0.42 秒交叉淡入
  - **七个角色各有自己的壁纸**（均从 HanaBi-ui-Genshin 导入）：甘雨、宵宫、八重神子、纳西妲、芙宁娜、珐露珊、钟离
  - 钟离的壁纸原项目只提供 `.hbw` 打包格式，这里从容器里提取图层**合成**出静态图（见「已知的坑」）
  - 七种标识画法：冰晶、焰火、雷光、岩印、草叶、水珠、风纹
  - **壁纸染色在着色器里做**（按元素色重映射亮度），不是叠半透明色层——后者会把暗部提亮、糊掉细节
  - **压暗强度按壁纸实测亮度逐个给**：甘雨 0.27 用 0.32，八重/纳西妲 0.80 用 0.66；一个统一值必然有一半不合适
  - 头像：有素材用角色头像（甘雨），没有就用元素色圆 + 首字占位
  - 新增角色只改 `internal/theme/theme.go` 一条数据，页面与绘制层不动
- **甘雨专属细节**（对照 `ganyu_ui.cpp` 逐项补齐）：
  - 三个窗口按钮是**冰晶切角八边形 + 淡金线 + 冰晶化图标**，悬停 140ms 补间；
    关闭按钮 hover 转玫红
  - 顶栏正中是命之座 `S I N U S   U N I C O R N I S`，HanaBi 字样只在状态栏
  - 侧栏底部与菜单选中项是**清心纹**（不再是冰晶），内容区右下角一枚大淡花
  - 主题卡片：清心纹 + 三条风纹 + 中国结 + 放大到 (36,31) 的冰元素徽记
- **其余六个角色的主题效果**（对照各自 `_ui.cpp` 逐项补）：
  - **窗口按钮**：宵宫暖金圆角 + 顶部金线；八重紫调 + 金弧 + 菱形；纳西妲深绿大圆角
    + 叶弧；芙宁娜蓝调 + 银弧；珐露珊切角 + 三角；钟离切角 + 底线
  - **侧栏底部与导航选中项**：金鱼/烟花、樱花、叶枝、礼帽、书与棱晶、岩印
  - **主题卡片**：每个角色一套纹样组合（金鱼+烟花、樱花+御守、叶枝、礼帽+水波、
    棱晶+古书、岩印+回纹），元素徽记统一放大到 (36,31)
  - **内容区右下角**：对应的大幅淡纹
- **开场分镜**（按 `Opening` 分发，时间轴仍是 6.2 秒）：
  - 甘雨：霜莲珠升起 + 三条冰绸 + 70 粒冰星迸发 + 2.42s 闪光
  - 宵宫：升空拖尾 + 主/副烟花 + 两只金鱼环绕 + 火舌标识
  - 八重神子：三座神樱 + 雷弧连到中心 + 落樱 + 雷光标识
  - 纳西妲：莲台绽放 + 梦境拱殿 + 叶环与光尘 + 草叶标识
  - 芙宁娜：两侧幕布拉开 + 三层水上舞台 + 落珠/水花 + 水滴标识
  - 珐露珊：六叶机关旋转展开 + 三道风带 + 光尘 + 风纹标识
  - 钟离：八枚岩方块下坠散开 + 石柱厅 + 金菱粒子 + 岩印标识
  - 这是纯 2D 分镜；原版的开场标识是几千个三角形的 3D 材质网格，这里不逐顶点复刻
- **开场动画（6.20 秒）**：时间轴与粒子参数**照搬原项目**；
  **动画期间品牌位留空**，标识飞抵品牌图标位置后由静态标识接手（交接契约有像素级回归测试 `TestBrandHandoff`）
  - 时间轴：0.20s 上升 → 1.45s 转向 → 3.65s 成形 → 4.65s 停靠 → 6.20s 结束
  - 甘雨保留 32 片雪 + 20 盏孔明灯；其余角色用各自的背景建筑/光景与元素氛围粒子
  - 标识按当前角色的 Motif 画：冰晶 / 焰火 / 雷光 / 岩印 / 草叶（三片叶子 + 芽心）/
    水滴（水纹扩散）/ 风纹（回旋气流），不再把草、水、风回落到冰晶
  - 结束时标识收缩并停靠到侧栏头像位置，开场层淡出、主界面「透出来」（不切视图）
  - 点画面任意处或按 Esc 可跳过（0.9 秒内快进走完，不硬切）
- **动态壁纸**：`internal/wallpaper/wallpaper.kage`（kage 着色器）一次绘制完成
  「Cover 铺满 + 人物区域内的呼吸与发丝漂动 + 主题染色」，位移按图片空间计（窗口缩放观感一致）；
  `Play` 参数可关闭动效。
- **12fps 后台时钟**驱动壁纸（`Painter.AnimationFrameRate`，只重绘那一层、不重建元素树）
- **主题页双视图**：标题右侧的「主题库 / 外观与启动」只在页内切换，不再跳侧栏；
  这是外观与启动的唯一入口，侧栏「设置」不再放这些内容。
  - **主题库视图**：元素筛选 + 卡片网格
  - **外观与启动视图**：用 mygo 的 `Slider` / `Switch` / `Segmented` / `Button`
    - **整体字号** 100%–125%（默认 112%）：所有 UI 文字走 `App.fs()`，布局尺寸不变（与参考的 `Fs()` 同语义）
    - **壁纸显现** 0.2–1.0：叠加到角色自带的 `Scrim` 上，越低越压暗
    - **背景流动 / 启动时播放开场 / 流畅动效 / 换主题时重播开场** 四个开关
    - **下次启动**：跟随当前 / 指定主题 / 随机角色；指定时用一排可换行的角色胶囊，随机不含黑白主题
    - 设置写在 `PathUserData/settings.json`（原子替换）；拖动类改动 300ms 防抖，退出时再 `FlushSettings`
- **通用设置页**：侧栏「设置」现在只放应用级内容，不再和主题页抢同一页：
  - **版本与更新**：当前版本（`mygo.App.Version()`）；检查、下载、安装、重启的完整流程
    （开发/测试构建显示未启用，不会联网）
  - **数据与存储**：`settings.json` 路径、复制路径、恢复默认设置
  - **关于**：参考项目链接与免责声明
- **启动记忆**：`internal/settings` 负责默认值、夹取与 JSON 读写；`NewAppWithSettings` 在启动时
  按模式选角色、字号、动效与开场开关；切换主题时同步 `Theme`（跟随当前用）
- **转场**：
  - 主题切换 **0.42 秒交叉淡入**：旧壁纸留在底层，新壁纸作为普通子元素按 `k` 淡入，
    同时 `a.tok` 在两组令牌之间插值；UI 整层是 Absolute，排在新壁纸之后绘制
  - 页面切换 **0.22 秒**：内容从下方 10 DIP 滑入并淡入
  - 主题页两个视图切换 **0.18 秒**淡入；窗口按钮悬停 140ms 补间，关闭按钮转玫红
- **半透明浮层**：侧栏、标题栏、状态栏直接压在壁纸上（不用 glass 材质——它自带边缘高光，
  会在侧栏上下亮出两条白边，见「已知的坑」）
- **主题卡片是纯色底 + 角色纹样**：七张卡片色彩分明，每张一套自己的图形记忆点；
  缩略图缩到卡片大小只剩一团色块，所以没有采用
- 顶部渐变压深带（190 DIP 渐隐到透明），让大标题压在花哨壁纸上仍清晰
- 窗口 + 侧栏 / 内容两栏 + 整宽底部状态栏

未完成（见路线图）：

- 开场与标识都是**2D 矢量近似**：原版的 3D 材质网格、逐顶点着色和体积粒子没有 1:1 复刻，
  分镜的构图/节奏/元素已经按 `characters/*/*_intro.cpp` 对齐
- 发布：GitHub 仓库 `lark203/bestuav-hanabi` 已创建；用 `mygo build -upload` 把
  安装包、更新压缩包与 `update-windows-amd64.json` 上传为 draft release；发布后
  自动更新才能端到端生效
- 壁纸页按需求保持现状（占位）；黑白基础主题暂不做；最小化退场动画待定

## 预览

| 主题库 · 甘雨（冰） | 主题库 · 宵宫（火） | 主题库 · 八重神子（雷） |
|---|---|---|
| ![甘雨](docs/screenshots/theme-ganyu.png) | ![宵宫](docs/screenshots/theme-yoimiya.png) | ![八重神子](docs/screenshots/theme-yae.png) |

| 主题库 · 纳西妲（草） | 主题库 · 芙宁娜（水） | 主题库 · 珐露珊（风） |
|---|---|---|
| ![纳西妲](docs/screenshots/theme-nahida.png) | ![芙宁娜](docs/screenshots/theme-furina.png) | ![珐露珊](docs/screenshots/theme-faruzan.png) |

| 主题库 · 钟离（岩） |
|---|
| ![钟离](docs/screenshots/theme-zhongli.png) |

| 主题页 · 外观与启动 | 设置 · 通用 |
|---|---|
| ![外观与启动](docs/screenshots/appearance.png) | ![设置](docs/screenshots/settings.png) |

| 开场 · 宵宫（烟花金鱼） | 开场 · 纳西妲（梦境莲台） | 开场 · 钟离（岩柱流星） |
|---|---|---|
| ![宵宫开场](docs/screenshots/opening-yoimiya.png) | ![纳西妲开场](docs/screenshots/opening-nahida.png) | ![钟离开场](docs/screenshots/opening-zhongli.png) |

| 开场 · 八重神子（神樱雷） | 开场 · 芙宁娜（水上舞台） | 开场 · 珐露珊（机关风带） |
|---|---|---|
| ![八重开场](docs/screenshots/opening-yae.png) | ![芙宁娜开场](docs/screenshots/opening-furina.png) | ![珐露珊开场](docs/screenshots/opening-faruzan.png) |

| 开场 · 抽枝生长（0.9s） | 开场 · 转向成形（2.5s） | 开场 · 完全成形（3.75s） |
|---|---|---|
| ![rise](docs/screenshots/intro-1-rise.png) | ![turn](docs/screenshots/intro-2-turn.png) | ![formation](docs/screenshots/intro-3-formation.png) |

## 已知的坑（写下来免得重犯）

1. **`Painter.Line` 只能画水平或垂直线**。它的文档写明 "straight horizontal or vertical line"，实现是取两点包围盒再用 width 覆盖某一维——**斜线会被压成一个矩形**。冰晶六条斜臂全变成竖线、雪花变成小方块，就是这个原因。任意角度的线、贝塞尔曲线和闭合多边形用 `ui.Path` + `Painter.StrokePath/FillPath`（见 `ornament.go` 的清心纹、中国结、切角八边形）；`seg`（Transform + Fill）是每段画一个胶囊的老办法，折点会鼓包，只适合单段短线。
2. **`image.RGBA` 是预乘 alpha**。往 `SetRGBA` 里写 `(255,255,255,0)` 这种 RGB>A 的非法值，渲染器会把边缘也当成不透明——光晕变成实心圆盘。必须写 `(v,v,v,v)`。
3. **不能用多层同心圆模拟径向渐变**：每层都是硬边，叠出来是一圈圈台阶，比单层圆更难看。正确做法是程序化生成一张渐变贴图（`ui.NewBitmap`）再用 `ImageRegion` 拉伸 + `Tint`（见 `glow`）。
4. **壁纸染色不能靠叠半透明色层**：那样只会把暗部提亮、整体发灰，蓝紫色像素乘上低亮度几乎不变，换主题时看不出区别。要在着色器里**按元素色重映射亮度**（`luma / tintLuma`），明暗层次才保留。
5. **`Material` 会替换 `Background`**：`glass.Glass` 是材质，同时写 `Background` 是死代码；悬停想加色只能重新 `Material(Tint)`。
6. **离屏截图的画布默认是白色**：开场层底色带透明度，不先铺不透明底会得到"白茫茫"的假象（真实运行时下面压着主界面）。
7. **根容器必须 `Fill()` + `AlignItems(ui.Stretch)`**：只写 `Grow(1)` 时它只占父容器主轴（宽度），子列会按内容高度垂直居中。
8. **资源根不能用「找 assets 目录」来定位**：包目录 `internal/assets` 与资源目录同名，向上查找会误命中。现在优先用 `mygo.App.Path(mygo.PathResources)`，再回退 `go.mod` 锚点。
9. **图片解码必须缓存**：壁纸是 1200 万像素的 JPG，每帧解码会直接把帧率打到个位数（见 `internal/assets`）。
10. **CPU 渲染器跑 kage 很慢**：离屏截图走 CPU 解释器，1280×800 一帧要数秒（`internal/view` 的测试因此要跑一分钟左右）；真机走 GPU 才够 12fps 或 60fps。写交互测试要先用小尺寸/少帧数，别在 1280×800 里跑几十帧动画。
11. **资源不随仓库分发**：`assets/` 在 `.gitignore` 里，打包前必须在本地存在（`mygo.json` 的 `resources: ["assets"]` 会把它复制到 exe 旁）。
12. **窗口按钮不要用 `—` `□` `✕` 这些字符**：字体里不一定有这些字形，缺字形时按钮变成一个空方块（试过）。通用主题用 SVG（`icons.Minimize/Maximize/Close`），各角色的专属按钮在 `chrome.go` 里用 `ui.Path` 画。
13. **侧栏不要用 glass 材质**：`Clear` 在亮壁纸（钟离的金色、纳西妲的浅绿）上几乎不压暗，菜单文字直接糊掉；`Regular` 又自带边缘高光，上下亮出两条白边。现在直接 `Background(令牌色)`，通透感靠 alpha。
14. **`.hbw` 是打包格式，但里面内嵌了原始 PNG/JPEG**：文件头 `HBW0001` 之后跟一串图层，偶数层是图像、奇数层是位移/遮罩纹理（**不是 alpha**，按它叠加会得到一片空白）。直接用图像层自带的 alpha 依序叠在底色上就能还原成一张壁纸——钟离那张就是这么来的。
15. **固定 `Height` 里不要再用上下 `Padding`**：`Height(34).Padding(14,16)` 的内容盒只剩 6px，文字直接被裁。固定高度容器用 `Padding(0, x)` + `AlignItems(Center)`。
16. **`ui.Scroll` 只在父容器高度有界时才滚动**：中间隔一层 Column 时，那层必须 `Grow(1)`，否则它按内容自然长高、Scroll 拿到无限高度——表现为"最后一行被裁掉且滚轮无反应"（主题卡片网格踩过）。验证滚动用 `Tester.Scroll(x,y,dx,dy)` 模拟滚轮再比对像素。
17. **卡片 `Grow(1)` 在不满一行的行里会被拉通栏**：最后一行只有一张卡片时会占满整行宽度，要配 `MaxWidth` 兜住。
18. **「只在选中时出现的元素」要用透明占位**：导航选中项左侧的竖杠如果只在 active 时才加进子列表，选中项的图标会比其他行右移 3px。始终放一个 `Alpha(0)` 的占位，两种状态宽度一致。
19. **`Padding` 四值是 CSS 顺序（上右下左），写错槽位会静默失效**：把底部留白 46 写进了「左」槽位，结果文字贴着窗口底、整块内容右偏 14px——两个症状一个根因。用像素扫描测量真实位置（逐行找亮像素）比目测截图可靠得多。
20. **改资源路径时，写死的路径不会报错、只会静默变空**：把 `ganyu-icon.png` 改名成 `icon.png` 后，`assets.Bitmap` 返回 nil，头像是"消失了"而不是报错。判断这类问题要看布局：`Tester.Find("花火")` 返回的 x 若等于左内边距，说明前面的元素宽度为 0（根本没占位）。
21. **`Absolute` 子元素最后绘制**：mygo 先画普通子元素、再画 Absolute。主题交叉淡入时如果新壁纸做成 Absolute，会整屏盖住侧栏和卡片（试过，一片发白）。正确结构是「新壁纸普通子元素 + UI 整层 Absolute」（见 `app.go` 的 View）。反过来也成立：开场覆盖层是 Absolute，所以它自然压在 UI 上。
22. **`Element.Opacity` 会连着色器一起淡**：`shader.Draw` 文档写明 "clipped, transformed and faded with the painter"，所以壁纸交叉淡入不用给 kage 加 alpha uniform，`Opacity(k)` 就够。注意分层顺序（坑 21），否则淡入的是壁纸但盖住了 UI。
23. **过渡不要直接用 `time.Now` 写死**：离屏渲染一帧要几秒，真时钟会在一帧内把 0.42s 过渡走完，截图测不到中间态。`App.clock` 可注入（测试里固定时间），`themeProgress` / `pageProgress` 都从它取时间；生产默认 `time.Now`。
24. **测试里不要用 `NewAppWithSettings` 做会触发的改动**：它 `persist=true`，`setCharacter` 会立刻写真实用户目录。要测过渡用 `NewApp`（persist=false），要测启动选择用 `NewAppWithSettings` 但不改设置。
25. **`seg` 的圆角半径必须夹到段长的一半**：开场里一条曲线会放大到半径 200px，相邻采样点只有几像素。用 `width/2` 当圆角时，短段的圆角半径超过短边一半，渲染出来只剩两端的圆帽——叶子/气流变成一串点。现在 `arcLine` / `quadLine` 直接用 `ui.Path` + `StrokePath` 一次描边（连续、带圆角连接），`seg` 也把半径夹到 `l/2`；短装饰线才继续用 `seg`。

## 运行

```bash
# 开发（热重载）
go tool mygo dev

# 直接跑
go run .

# 编译校验
go build ./...
```

## 打包与自动更新

`mygo.json` 的 `updates.publicKey` 是更新签名公钥，`updates.github` 指向
`lark203/bestuav-hanabi`：`mygo build -upload` 会把安装包、更新压缩包与
`update-windows-amd64.json` 作为 GitHub draft release 上传，发布后应用从
GitHub 的 latest release 读取 manifest。私钥**不进仓库**，放在仓库外的
`D:\workbuddy\bestuav-hanabi-update-keys\mygo-update.key`（或 CI secret），
构建/上传时通过 `MYGO_UPDATER_PRIVATE_KEY` 和 `GH_TOKEN` 传入：

```powershell
$env:MYGO_UPDATER_PRIVATE_KEY = (Get-Content -Raw D:\workbuddy\bestuav-hanabi-update-keys\mygo-update.key).Trim()
$env:GH_TOKEN = "ghp_..."           # 上传时需要
go tool mygo build -upload
# 或者
.\scripts\build-release.ps1 -Upload
```

产物在 `build/windows-amd64/`：`花火.exe`、`hanabi-<version>-windows-amd64.tar.gz`、
`hanabi-setup-<version>-windows-amd64.exe`，以及 `update-windows-amd64.json`。
把压缩包与 manifest 上传到 GitHub draft release（`-upload` 会自动做，文件名保持不变），
发布 release 后，已安装的旧版本就能在「设置 → 版本与更新」里检查、下载、安装并重启。

- 自动更新的检查和安装都在 goroutine 里，UI 只读原子进度；开发/测试构建
  `Updater.Enabled()=false`，不会联网。
- `go tool mygo build` 在没有私钥时会跳过更新文件；发布版必须带私钥构建。
- 当前 GitHub 仓库已创建但还没有 release；首次发布需要 `GH_TOKEN`，
  并用 `mygo build -upload` 生成 draft，再在 GitHub 上发布。

## 自动构建（GitHub Actions）

详细步骤见 [docs/github-release.md](docs/github-release.md)。

`.github/workflows/release.yml` 在推送 `v*` tag 时创建 draft release，按
Windows / macOS / Linux 三个 runner 串行构建并上传安装包、更新压缩包和 manifest。
原生 UI 不需要 Bun；Windows 安装包在 Windows runner 上生成。

**必须配置的 secret**：

```bash
# 更新签名私钥（仓库外的 mygo-update.key）
gh secret set MYGO_UPDATER_PRIVATE_KEY < D:/workbuddy/bestuav-hanabi-update-keys/mygo-update.key
```

`GH_TOKEN` 由 workflow 的 `github.token` 提供，不需要手动配置；`permissions.contents: write`
已经写在 workflow 里。**不要把私钥提交到仓库**。

GitHub 仓库还需要打开写权限：**Settings → Actions → General → Workflow permissions →
Read and write → Save**（默认只读时 `gh release create/upload` 会 403）。

发布流程：

```bash
# mygo.json 的 version 与 tag 必须一致，例如 0.1.1 / v0.1.1
git tag v0.1.1
git push github v0.1.1
```

只有推 GitHub 的 tag 会触发 CI；推 AtomGit 不会。在 Actions 里等三个平台跑完，
检查 draft release 并发布。发布后旧版本就能自动更新。

### 本机推送与代理

本机 WorkBuddy 会注入 `HTTP_PROXY=127.0.0.1:5437`，它会拦 `github.com`；
按 bestuav-mygo 的记录，给 github.com 单独走 Clash 的 7890：

```bash
git config --global http.https://github.com.proxy http://127.0.0.1:7890
git remote add github https://github.com/lark203/bestuav-hanabi.git
git push github main dev
```

回滚：`git config --global --unset http.https://github.com.proxy`。
凭证沿用本机 Git Credential Manager 里 bestuav-mygo 同一套 GitHub 凭证即可。

可选、当前**没有配置**的签名：

- Windows 代码签名：需要 `WINDOWS_CERTIFICATE`（base64 pfx）和
  `WINDOWS_CERTIFICATE_PASSWORD`，并在 workflow 里加 osslsigncode 步骤；不加也能构建，
  但 SmartScreen 会提示未知发布者。
- macOS Developer ID / notarization：需要 `MACOS_CERTIFICATE`、密码、`APPLE_ID`、
  `APPLE_TEAM_ID`、`APPLE_APP_PASSWORD`；不加则是 ad hoc 签名，别的 Mac 上
  Gatekeeper 会拦截。

`assets/` 已随仓库提交（约 18MB），runner 不需要额外下载壁纸。

## 目录

```text
main.go                 启动：窗口 + Content
mygo.json               打包配置（name/slug/version；resources 把 assets/ 复制到 exe 旁；
                        updates 更新公钥与发布目录）
CHANGELOG.md            发布说明（mygo build 写进 update manifest）
scripts/build-release.ps1 带更新私钥的发布构建
internal/theme/         角色包：令牌 + 元素 + 标识画法 + 壁纸染色强度
internal/assets/        资源定位（resources 目录 / go.mod 锚点）+ 解码缓存
internal/wallpaper/     wallpaper.kage + Draw：动态壁纸着色器
internal/intro/         开场的时间轴、粒子与标识姿态（纯逻辑，可单测）
internal/settings/      跨启动设置：默认值、夹取、settings.json 读写
internal/ui/icons/      内联 SVG 图标（currentColor 跟随主题）
internal/view/          界面根：App 状态、开场渲染、侧栏、标题栏、页面
                        ornament.go：清心纹 / 中国结 / 切角八边形（ui.Path）
                        ornament_primitives.go：金鱼/樱花/叶枝/礼帽/棱晶/古书/岩印
                        ornaments_ui.go：纹样在内容区/侧栏/导航/卡片上的排布
                        chrome.go：七个角色的窗口按钮造型
                        opening.go：七个角色的开场分镜与背景
                        appearance.go：主题页的外观与启动视图
                        settings_page.go：通用设置（版本/更新、数据、关于）
                        update.go：自动更新的检查/下载/安装状态机
docs/screenshots/       界面截图（离屏渲染产出）
```

## 测试

```bash
go test ./...          # 时间轴/粒子/姿态 + 开场渲染冒烟
```

`internal/intro` 的测试钉住时间轴（6.20s 结束、dt 夹到 0.20）、姿态起止点（中央偏下 → 侧栏停靠点，半径 14.4 = 品牌 glyph）、
粒子确定性（同一尺寸两次初始化完全一致）、跳过耗时（≤0.9s）以及淡出/交接的单调性。
`internal/wallpaper` 钉住"时间推进会改变画面、`play=0` 时完全静止"。
`internal/settings` 钉住默认值（112%、跟随当前）、越界夹取、坏 JSON 回退与读写往返。
`internal/theme` 钉住每个角色的 Chrome / Ornament 配置。
`internal/view` 另外钉住启动模式（跟随/固定/随机）、字号换算、主题过渡、页面入场、
主题页视图切换与更新检查的文本/进度；离屏渲染仍是主要成本（`go test ./...` 约一分半）。

## 路线图

1. **发布第一个 GitHub release**：在 `lark203/bestuav-hanabi` 用 `mygo build -upload`
   生成 draft release，验证「检查 → 下载 → 安装 → 重启」整条链路
2. **可选**：标识/开场的 3D 材质网格、最小化退场动画、更多分镜细节
3. 壁纸页保持现状；黑白基础主题不做

## 说明

- 角色名称与相关视觉元素归原版权方所有，本项目仅作技术学习与本地演示，不用于商业分发。
- 资产（壁纸、序列帧）不进入版本库，放在 `assets/`（已 gitignore），由本地脚本从原项目导入。
