#!/bin/bash

################################################################################
# Script Name : images.sh
# Description : Manage container images (pull/save/load/push) with flexible modes
# Author      : zong xun
# Version     : v1.0
# Created     : 2025-06-30
# Updated     : 2025-06-30
################################################################################

# 初始化变量
BASE_DIR="$( dirname "$( readlink -f "${0}" )" )"
INSTALL_LOG_PATH="${BASE_DIR}/images.log"
images_dir="${BASE_DIR}/images"
http_proxy=${http_proxy:-"http://192.168.21.101:7890"}
https_proxy=${https_proxy:-"http://192.168.21.101:7890"}
HTTP_PROXY="$http_proxy"
HTTPS_PROXY="$https_proxy"

# 日志函数
log() {
  local level="$1"; shift
  local color_reset="\033[0m"
  local color_info="\033[1;34m"
  local color_warn="\033[1;33m"
  local color_error="\033[1;31m"
  local color_ok="\033[1;32m"
  local now
  now=$(date +'%Y-%m-%dT%H:%M:%S%z')
  case "$level" in
    INFO)  echo -e "${color_info}[Info][$now]: $*${color_reset}" | tee -a "${INSTALL_LOG_PATH}" ;;
    WARN)  echo -e "${color_warn}[Warn][$now]: $*${color_reset}" | tee -a "${INSTALL_LOG_PATH}" ;;
    ERROR) echo -e "${color_error}[Error][$now]: $*${color_reset}" | tee -a "${INSTALL_LOG_PATH}" ; exit 1 ;;
    OK)    echo -e "${color_ok}[OK][$now]: $*${color_reset}" | tee -a "${INSTALL_LOG_PATH}" ;;
    *)     echo "[Log][$now]: $*" | tee -a "${INSTALL_LOG_PATH}" ;;
  esac
}

info()  { log INFO  "$@"; }
warn()  { log WARN  "$@"; }
error() { log ERROR "$@"; }
ok()    { log OK    "$@"; }

installed() {
  command -v "$1" >/dev/null 2>&1
}

init_log() {
  touch "${INSTALL_LOG_PATH}" || error "Failed to create log file ${INSTALL_LOG_PATH}"
  info "Log file created at path: ${INSTALL_LOG_PATH}"
}

# 参数帮助
usage() {
  echo "Usage: $0 -a <action> -f <image_list> [-r <registry>] [-p <project>] [--push-mode <1|2|3>] [--load-mode <1|2|3>] [--save-mode <1|2|3>]"
  echo
  echo "Actions:"
  echo "  pull            Pull images from external registry"
  echo "  save            Save images into tar files"
  echo "  load            Load images from tar files"
  echo "  push            Push images to private registry"
  echo
  echo "Options:"
  echo "  -a              Action (required): pull | save | load | push"
  echo "  -f              Image list file (required for pull/save/push)"
  echo "  -r              Target registry (default: registry.k8s.local)"
  echo "  -p              Target project namespace (default: base)"
  echo
  echo "Modes:"
  echo "  --push-mode     Push mode (default: 1):"
  echo "                    1 = Push as registry/image:tag"
  echo "                    2 = Push as registry/project/image:tag"
  echo "                    3 = Push preserving original project path"
  echo
  echo "  --load-mode     Load mode (default: 1):"
  echo "                    1 = Load all *.tar files from current directory"
  echo "                    2 = Load all *.tar files from ./images directory"
  echo "                    3 = Recursively load *.tar from subdirectories under ./images/*/"
  echo
  echo "  --save-mode     Save mode (default: 1):"
  echo "                    1 = Save tar files to current directory"
  echo "                    2 = Save tar files to ./images/"
  echo "                    3 = Save tar files to ./images/<project>/"
  exit 1
}


# 先检查 getopt 是否支持长选项
getopt --test > /dev/null
if [[ $? -ne 4 ]]; then
  echo "Error: Enhanced getopt is not supported on this system."
  exit 1
fi

# 用 getopt 解析参数
PARSED=$(getopt -o a:f:r:p: --long action:,file:,registry:,project:,push-mode:,load-mode:,save-mode: -- "$@")
if [[ $? -ne 0 ]]; then
  usage
fi

eval set -- "$PARSED"

while true; do
  case "$1" in
    -a|--action)       action="$2"; shift 2;;
    -f|--file)         images_list="$2"; shift 2;;
    -r|--registry)     registry_name="$2"; shift 2;;
    -p|--project)      project="$2"; shift 2;;
    --push-mode)       push_mode="$2"; shift 2;;
    --load-mode)       load_mode="$2"; shift 2;;
    --save-mode)       save_mode="$2"; shift 2;;
    --) shift; break;;
    *) usage;;
  esac
done

[ -z "$action" ] && error "Missing required -a <action> parameter" && usage


registry_name=${registry_name:-'registry.k8s.local'}
project=${project:-'library'}

init_log
installed docker || installed podman || installed nerdctl || error "docker, podman or nerdctl not found."

docker=$(command -v docker || command -v podman || command -v nerdctl)
if [ "$(basename "$docker")" = "nerdctl" ]; then
  docker="/usr/local/bin/nerdctl --insecure-registry"
fi

# load 分模式
image_load_mode_1() {
  > images_load.txt
  for i in *.tar; do
    sudo $docker load -i "$i" | tee -a images_load.txt
  done
}

image_load_mode_2() {
  for obj in "$images_dir"/*.tar; do
    sudo $docker load -i "$obj"
  done
}

image_load_mode_3() {
  for obj in $(ls "${BASE_DIR}"); do
    if [ -d "${BASE_DIR}/$obj" ]; then
      for image in $(find "${BASE_DIR}/$obj/images" -name '*.tar' 2>/dev/null); do
        sudo $docker load -i "$image" | tee -a images_load.txt
      done
    fi
  done
}

# save 分模式
# 保存到当前目录（当前 shell 的 pwd）
image_save_mode_1() {
  while read -r line; do
    image_repo=$(echo "$line" | awk -F '/' '{print $1}')
    image_name=$(echo "$line" | awk -F '/' '{print $NF}' | awk -F ':' '{print $1}')
    image_tag=$(echo "$line" | awk -F '/' '{print $NF}' | awk -F ':' '{print $2}')
    sudo $docker save -o "./${image_repo}_${image_name}_${image_tag}.tar" "$line"
  done < "$images_list"
}

# 保存到 ${images_dir}
image_save_mode_2() {
  mkdir -p "${images_dir}"
  while read -r line; do
    image_repo=$(echo "$line" | awk -F '/' '{print $1}')
    image_name=$(echo "$line" | awk -F '/' '{print $NF}' | awk -F ':' '{print $1}')
    image_tag=$(echo "$line" | awk -F '/' '{print $NF}' | awk -F ':' '{print $2}')
    sudo $docker save -o "${images_dir}/${image_repo}_${image_name}_${image_tag}.tar" "$line"
  done < "$images_list"
}

# 保存到 ${images_dir}/${image_project}
image_save_mode_3() {
  while read -r line; do
    image_repo=$(echo "$line" | awk -F '/' '{print $1}')
    image_project=$(echo "$line" | awk -F '/' '{print $(NF-1)}')
    image_name=$(echo "$line" | awk -F '/' '{print $NF}' | awk -F ':' '{print $1}')
    image_tag=$(echo "$line" | awk -F '/' '{print $NF}' | awk -F ':' '{print $2}')
    mkdir -p "${images_dir}/${image_project}"
    sudo $docker save -o "${images_dir}/${image_project}/${image_repo}_${image_project}_${image_name}_${image_tag}.tar" "$line"
  done < "$images_list"
}

# push 分模式
image_push_mode_1() {
  while read -r line; do
    image_name=$(echo "$line" | awk -F '/' '{print $NF}' | awk -F ':' '{print $1}')
    image_tag=$(echo "$line" | awk -F '/' '{print $NF}' | awk -F ':' '{print $2}')
    sudo $docker tag "$line" "${registry_name}/${image_name}:${image_tag}"
    sudo $docker push "${registry_name}/${image_name}:${image_tag}"
  done < "$images_list"
}

image_push_mode_2() {
  while read -r line; do
    image_name=$(echo "$line" | awk -F '/' '{print $NF}' | awk -F ':' '{print $1}')
    image_tag=$(echo "$line" | awk -F '/' '{print $NF}' | awk -F ':' '{print $2}')
    sudo $docker tag "$line" "${registry_name}/${project}/${image_name}:${image_tag}"
    sudo $docker push "${registry_name}/${project}/${image_name}:${image_tag}"
  done < "$images_list"
}

image_push_mode_3() {
  while read -r line; do
    image_project=$(echo "$line" | awk -F '/' '{print $(NF-1)}')
    image_name=$(echo "$line" | awk -F '/' '{print $NF}' | awk -F ':' '{print $1}')
    image_tag=$(echo "$line" | awk -F '/' '{print $NF}' | awk -F ':' '{print $2}')
    sudo $docker tag "$line" "${registry_name}/${image_project}/${image_name}:${image_tag}"
    sudo $docker push "${registry_name}/${image_project}/${image_name}:${image_tag}"
  done < "$images_list"
}

# pull/save/load/push 调度
image_pull() {
  while read -r line; do
    sudo env http_proxy=$http_proxy https_proxy=$https_proxy $docker pull "$line"
  done < "$images_list"
}

image_save() {
  case "$save_mode" in
    2) image_save_mode_2;;
    3) image_save_mode_3;;
    *) image_save_mode_1;;
  esac
}

image_load() {
  case "$load_mode" in
    2) image_load_mode_2;;
    3) image_load_mode_3;;
    *) image_load_mode_1;;
  esac
}

image_push() {
  case "$push_mode" in
    2) image_push_mode_2;;
    3) image_push_mode_3;;
    *) image_push_mode_1;;
  esac
}

# 主调度
case "$action" in 
   pull)  image_pull;;
   save)  image_save;;
   load)  image_load;;
   push)  image_push;;
   *)     usage;;
esac

exit 0

