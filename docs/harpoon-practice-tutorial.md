# 一叉子搞定镜像搬运：Harpoon 使用实践教程

> 当你拉不完的镜像、打不完的包、传不完的仓库……别慌，先扔个 Harpoon 过去。

临近春节，不少人已经在回家路上了。这篇小文写给还在工位摸鱼、或者打算路上/过年抽空折腾一下镜像搬运的你——用 Harpoon 把拉、存、推一条龙串起来，省下的时间正好多陪家人。  
先祝各位**春节快乐，一路顺风**；下面正经聊工具。

---

## 一、我为什么要做这个项目？

做运维这些年，见过太多「企业离线/内网环境」的现场：  
客户机房没外网，要上一套 K8s、要部署一堆产品，**镜像从哪来？**  
答案永远是：先在能上网的机器上**拉**，再**打成 tar 包**，U 盘/专线拷过去，在目标环境** load**，再**推到客户自己的 Harbor**。  
拉、打包、拷贝、加载、推送、有时还要**清理**旧镜像……一套流程下来，命令行敲到怀疑人生，脚本写了好几版，还得分 Docker / Podman / 无 daemon 的 Skopeo 各自照顾一遍。

于是就有了一个朴素的想法：**能不能有一个小工具，认一份「镜像列表」，就能把「拉 → 存 → 拷 → 装 → 推」这条链路串起来？**  
而且最好别绑死某一家运行时，Docker、Podman、Nerdctl、Skopeo 都能用；支持批量、支持配置、支持 CI/CD，脚本里调用也舒服。

这就是 **Harpoon（hpn）** 的由来：像鱼叉一样，精准地「叉」住你要的那一堆镜像，该拉拉、该存存、该推推，一气呵成。

---

## 二、顺便聊聊：AI 时代，运维也能搞「Agent 开发」

这两年 AI 写代码、Copilot、Agent 已经不算新鲜事了。  
对运维来说，除了用 AI 查文档、写脚本，其实还可以再进一步：**自己设计小工具、小 Agent，把重复劳动交给程序。**

比如：

- 定好「镜像列表 + 目标仓库」这种简单输入；
- 工具自动选运行时、拉镜像、打包、推仓库，失败重试、日志清晰；
- 你在 CI 里调一条命令，或在本机敲一条命令，就完成以前要写一坨脚本才能做的事。

Harpoon 就是这种思路下的一个具体实现：**不替代 Docker/Podman，而是站在它们之上，把「批量镜像搬运」抽象成一套统一命令。**  
你完全可以在此基础上，再结合自己的场景（例如从 Helm 抽镜像列表、按环境推不同项目）做封装，甚至和 AI/Agent 结合，让「自然语言描述需求 → 生成镜像列表 → 调用 hpn」变成一条龙。  
运维不再只是「执行者」，也可以是「设计小工具、小流水线」的人。

---

## 三、Harpoon 能干啥？先有个全景图

一句话：**用一份「镜像列表文件」+ 几条命令，完成：拉取、保存成 tar、加载、推送到仓库、从 Helm 抽镜像列表，以及列出/清理本地镜像。**

支持多种容器运行时（Docker / Podman / Nerdctl / Skopeo），自动检测或指定；支持配置文件、环境变量、CI 友好（如 `--password-stdin`）。  
下面从「零基础」到「进 CI / 进生产」的顺序，一步步说怎么用。

---

## 四、安装与「你好世界」

### 4.1 安装

Linux / macOS 一条龙（以 Linux amd64 为例）：

```bash
curl -L https://github.com/Ghostwritten/harpoon/releases/latest/download/hpn-linux-amd64 -o hpn
chmod +x hpn
sudo mv hpn /usr/local/bin/
```

其他架构或 Windows 去 [Releases](https://github.com/Ghostwritten/harpoon/releases) 里挑对应二进制即可。  
验证：

```bash
hpn --version
```

看到版本号就说明安装成功。

### 4.2 最小闭环：拉两个镜像试试

先搞一个「镜像列表」文件，一行一个镜像名（支持 `#` 注释和空行）：

```bash
echo "nginx:latest" > images.txt
echo "alpine:3.18" >> images.txt
```

然后：

```bash
hpn pull -f images.txt
```

此时会用你本机已有的运行时（Docker / Podman / Nerdctl 等）自动拉取这两个镜像。  
**-f** 表示「从文件读镜像列表」，这是 Harpoon 里最常用的参数之一。

---

## 五、核心场景一：离线/内网搬运（拉 → 存 → 拷 → 装 → 推）

这是 Harpoon 的主战场：**有网的机器**上拉镜像、打成 tar，拷到**无网/内网**机器，再 load、再推到客户 Harbor。

### 5.1 在有网的机器上：拉取 + 保存为 tar

```bash
# 镜像列表（可以很多行）
cat > images.txt << EOF
nginx:1.21-alpine
redis:7-alpine
postgres:15-alpine
EOF

# 拉取
hpn pull -f images.txt

# 保存到当前目录下的 images/（默认就是 ./images）
hpn save -f images.txt --path ./images
```

若**网速够快**或机器资源充足，可以加上 `--workers` 控制并发数，多进程同时拉/存，整体会快不少（见下面第八节「并发数」）。

`./images` 下会按镜像名生成一堆 `xxx.tar`（以及可选校验和）。  
你可以把整个 `images/` 目录打包（如 tar.gz）拷走。

### 5.2 断点续传：save 中断了？再跑一遍就行

大批量 `hpn save` 时网络断了、磁盘满了、手滑 Ctrl+C 了……不用从头再来。  
Harpoon 的 **save 默认带断点续传**：每次保存都会在 tar 旁生成 `.sha256` 校验文件；下次再执行**同一条** `hpn save -f images.txt --path ./images` 时，会先检查目标目录里是否已有对应 `xxx.tar` 和 `xxx.tar.sha256`，若存在且校验通过，就**跳过**该镜像，只处理还没保存完的。

你看到的提示大概是：`⏭️ Skipped nginx:latest (already exists and verified)`。  
所以流程就是：**第一次能保存多少算多少，失败了再跑一遍同一句命令**，已完成的自动跳过，省时省心。

### 5.3 在离线/内网机器上：加载 + 推到自己的仓库

```bash
# 解压你拷过来的包后，加载镜像
hpn load --path ./images

# 推送到客户 Harbor（按原路径推送）
hpn push -f images.txt --registry harbor.company.com

# 或者全部推到同一项目下（统一项目名）
hpn push -f images.txt --registry harbor.company.com --project production
```

这样，**同一份 images.txt** 就贯穿了：拉、存、拷、装、推，全程可脚本化。

### 5.4 登录私有仓库

推送前如果仓库要认证：

```bash
# 交互式（最安全，密码不落命令行）
hpn login harbor.company.com

# 或直接传参（脚本/CI 里常用）
hpn login harbor.company.com -u admin -p yourpassword

# CI 里用 stdin 传密码
echo "yourpassword" | hpn login harbor.company.com -u admin --password-stdin
```

HTTP 或自签名证书的仓库可以加 `--insecure`。

---

## 六、核心场景二：镜像列表从哪来？Helm Chart 里「叉」出来

很多现场是「要上一套基于 Helm 的产品」，镜像全在 Chart 里。  
Harpoon 支持从 Helm Chart **直接抽出镜像列表**，再交给 pull/save/push，不用自己对着 values 一个个抄。

前提：本机已安装 [Helm](https://helm.sh) CLI。

```bash
# 从远程 chart 指定版本，导出镜像列表到 images.txt
hpn extract --chart bitnami/nginx --version 15.0.0 -o images.txt

# 然后照常拉、存、推
hpn pull -f images.txt
hpn save -f images.txt --path ./images
```

也支持本地 chart 目录或 tgz：

```bash
hpn extract --chart ./mychart -o images.txt
hpn extract --chart ./mychart-1.0.0.tgz -f values-prod.yaml -o images.txt
```

`-f values-prod.yaml` 会传给 `helm template`，这样渲染出来的镜像会和你要部署的环境一致。  
输出是一行一个镜像，和前面的 `images.txt` 格式完全一致，直接给 `hpn pull -f` / `hpn push -f` 用即可。

---

## 七、核心场景三：看清家里有啥——列镜像、做检查

Harpoon 还能「看看本地到底有哪些镜像 / 哪些 tar」，方便你做核对或清理。相比直接敲 `docker images` / `podman images` 一长串，**hpn ls** 的输出更规整、易读，还带统计。

### 7.1 列出当前运行时里的镜像

```bash
hpn ls
```

会用当前检测到的运行时（Docker/Podman/Nerdctl）列镜像，表头 + 分隔线 + 一行一个「镜像:标签」，最后给个总数。示例（节选）：

```
Listing images from docker

IMAGE                                             
--------------------------------------------------
alpine:latest
nginx:latest
redis:7-alpine
grafana/grafana:12.3.1
...
42 images
```

比传统 `docker images` 的宽表更省眼，脚本里也方便 parse。

### 7.2 列出某个目录下的 tar 文件（带大小、校验和）

```bash
hpn ls --path ./images
```

适合在「保存了一堆 tar」之后，确认是否齐全。输出会带 **SIZE** 和 **CHECKSUM**（是否已有 `.sha256` 且校验通过），一目了然：

```
IMAGE                                                SIZE  CHECKSUM
-------------------------------------------------------------------
docker.io/bats/bats:v1.4.1                       50.9 MiB  ok
docker.io/grafana/grafana:12.3.1                678.0 MiB  ok
ghcr.io/jkroepke/kube-webhook-certgen:1.7.4      23.6 MiB  ok

3 images
```

既好看又实用，比自己在目录里 `ls -lh *.tar` 再对校验和省事多了。

### 7.3 用列表文件做「一致性检查」

```bash
hpn ls -f images.txt
```

会检查 `images.txt` 里的镜像是否都在当前运行时里存在；也可以和 `--path ./images` 一起用，检查是否都已有对应 tar。  
脚本里可以用退出码判断（0 成功，1 运行时错误，2 用法错误），方便做自动化检查。

### 7.4 按列表删本地镜像（慎用）

```bash
hpn rmi -f images.txt
```

会按文件里的列表从本地运行时删镜像。需要强制删除时（如 Docker），可以传透传参数：  
`hpn rmi -f images.txt -- -f`。

---

## 八、用好配置和运行时，少敲键盘

### 8.1 配置文件

把常用 registry、项目、路径、运行时写进配置，以后很多参数就不用每次敲了。

```bash
mkdir -p ~/.hpn
cat > ~/.hpn/config.yaml << EOF
registry: harbor.company.com
project: production
runtime:
  preferred: docker
  auto_fallback: true
paths:
  save_path: ./images
  load_path: ./images
EOF
```

之后直接：

```bash
hpn pull -f images.txt
hpn save -f images.txt
hpn push -f images.txt
```

会按配置里的路径和仓库来。

### 8.2 指定运行时 / 自动回退

```bash
# 强制用 Podman
hpn --runtime podman pull -f images.txt

# CI 里希望「有谁用谁」
hpn --auto-fallback pull -f images.txt
```

在 GitLab CI / GitHub Actions 等流水线里经常用 `--auto-fallback`，避免「只有 Podman 没有 Docker」就报错。

### 8.3 透传参数给底层运行时

有些场景需要给 docker/podman/skopeo 传额外参数（比如跳过 TLS 校验）：

```bash
hpn pull -f images.txt -- --tls-verify=false
hpn push -f images.txt --registry harbor.company.com -- --tls-verify=false
```

`--` 后面的内容会原样传给底层命令，具体能传啥要看各运行时文档。

### 8.4 并发数（--workers）：网速快就多开几路

pull、save、load、push 都支持**多任务并行**。默认是 5 个 worker；若你网速够快、CPU/内存也跟得上，可以手动把并发数调高，一批镜像会明显拉得更快。

```bash
# 拉取时开 10 个并发
hpn pull -f images.txt --workers 10

# 保存、推送同理
hpn save -f images.txt --path ./images --workers 8
hpn push -f images.txt --registry harbor.company.com --workers 6
```

优先级：命令行 `--workers` &gt; 配置文件里的 `parallel.max_workers` &gt; 默认 5。  
网络一般或机器一般时，用默认或 `--workers 3` 即可；千兆/万兆、SSD、内存足，可以试试 10～15，按实际压测微调。

---

## 九、在 CI/CD 里用 Harpoon

思路都是：**镜像列表固定（或由 Helm extract 生成）→ 一条 hpn 命令完成拉/推。** 下面用当下主流的 GitLab CI、GitHub Actions 举例；其他如 Tekton、Argo Workflows 等也是同样用法，在流水线里装好 hpn 后调同样命令即可。

### 9.1 GitHub Actions

```yaml
- name: Deploy images
  run: |
    curl -sL -o hpn https://github.com/Ghostwritten/harpoon/releases/latest/download/hpn-linux-amd64
    chmod +x hpn && sudo mv hpn /usr/local/bin/
    echo "${{ secrets.REGISTRY_PASSWORD }}" | hpn login harbor.company.com -u "${{ secrets.REGISTRY_USER }}" --password-stdin
    hpn --auto-fallback push -f production-images.txt --registry harbor.company.com --project production
```

### 9.2 GitLab CI

在 `.gitlab-ci.yml` 里用一条 job 拉取并推送到私有仓库，例如：

```yaml
deploy-images:
  image: docker:24
  services: [docker:dind]
  variables:
    REGISTRY: harbor.company.com
    PROJECT: production
  before_script:
    - apk add --no-cache curl
    - curl -sL -o /usr/local/bin/hpn https://github.com/Ghostwritten/harpoon/releases/latest/download/hpn-linux-amd64
    - chmod +x /usr/local/bin/hpn
    - echo "$REGISTRY_PASSWORD" | hpn login $REGISTRY -u "$REGISTRY_USER" --password-stdin
  script:
    - hpn --auto-fallback push -f production-images.txt --registry $REGISTRY --project $PROJECT
  only:
    - main
```

（其中 `REGISTRY_USER` / `REGISTRY_PASSWORD` 放在 GitLab CI/CD Variables 里，并勾选 Masked。）

### 9.3 其他流水线（Tekton、Argo 等）

拉 → 存 → 归档 → 部署阶段再 load / push，和上面「离线搬运」逻辑一致：在对应步骤里安装 hpn（或使用带 hpn 的镜像），把 `images.txt` 和 `--registry` / `--project` 换成流水线变量即可。

---

## 十、小结：什么时候该想起 Harpoon？

- **企业离线/内网交付**：拉、打包、拷贝、load、推，一条链路用同一份列表和少量命令搞定。  
- **断点续传**：`hpn save` 会为每个 tar 生成 `.sha256`，再次执行同一命令时自动跳过已存在且校验通过的镜像，中断后重跑即可续传。  
- **并发数**：网速或机器资源够用时可加 `--workers N`（pull/save/load/push 均支持），多进程并行，加快整批操作。  
- **从 Helm 出镜像列表**：`hpn extract` 一把梭，再 pull/save/push。  
- **多运行时环境**：Docker / Podman / Nerdctl / Skopeo 统一用一套命令，配合 `--auto-fallback` 更省心。  
- **CI/CD**：GitHub Actions、GitLab CI 等主流流水线里，登录 + 一条 push，配合 `--password-stdin` 和退出码，脚本友好。  
- **日常盘点与清理**：`hpn ls` / `hpn ls --path` / `hpn ls -f` / `hpn rmi -f`，看清家里有啥、按列表删。

如果你也经常在「一堆镜像要拉、要存、要推、要清理」里打转，不妨试试 Harpoon，把重复劳动交给这一把小「鱼叉」，自己省出时间搞点更香的事——比如再写个小 Agent 自动生成 `images.txt`。

- **项目地址**: [GitHub - Ghostwritten/harpoon](https://github.com/Ghostwritten/harpoon)  
- **文档**: 仓库内 [docs/](https://github.com/Ghostwritten/harpoon/tree/main/docs)（Quick Start、Examples、Changelog 等）

祝镜像搬运愉快，一叉一个准。
