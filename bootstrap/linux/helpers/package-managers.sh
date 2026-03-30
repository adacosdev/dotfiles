#!/bin/bash

# Linux distro and package-manager primitives shared by bootstrap entrypoints.

linux_distro_id() {
  if [[ -n "${CHEZMOI_LINUX_DISTRO_ID:-}" ]]; then
    printf '%s\n' "$CHEZMOI_LINUX_DISTRO_ID"
    return
  fi

  if [[ -r /etc/os-release ]]; then
    # shellcheck disable=SC1091
    . /etc/os-release
    printf '%s\n' "${ID:-unknown}"
    return
  fi

  printf 'unknown\n'
}

is_apt_distro() {
  case "$(linux_distro_id)" in
    ubuntu|debian) return 0 ;;
    *) return 1 ;;
  esac
}

is_pacman_distro() {
  case "$(linux_distro_id)" in
    arch|endeavouros) return 0 ;;
    *) return 1 ;;
  esac
}

is_dnf_distro() {
  case "$(linux_distro_id)" in
    fedora) return 0 ;;
    *) return 1 ;;
  esac
}

pkg_update() {
  if is_apt_distro; then
    sudo apt update
    return
  fi

  if is_pacman_distro; then
    sudo pacman -Sy --noconfirm "$@"
    return
  fi

  if is_dnf_distro; then
    sudo dnf makecache
    return
  fi

  return 1
}

pkg_install() {
  if is_apt_distro; then
    sudo apt install -y "$@"
    return
  fi

  if is_pacman_distro; then
    sudo pacman -S --noconfirm "$@"
    return
  fi

  if is_dnf_distro; then
    sudo dnf install -y "$@"
    return
  fi

  return 1
}

enable_service_now() {
  sudo systemctl enable --now "$1"
}

add_user_to_group() {
  sudo usermod -aG "$1" "$USER"
}

install_docker_stack() {
  if is_apt_distro || is_pacman_distro; then
    pkg_install docker docker-compose
  elif is_dnf_distro; then
    pkg_install moby-engine docker-compose
  else
    return 1
  fi

  enable_service_now docker
  add_user_to_group docker
}

install_pyenv_build_dependencies() {
  if is_apt_distro; then
    pkg_update
    pkg_install build-essential libssl-dev zlib1g-dev libbz2-dev libreadline-dev libsqlite3-dev curl git libncursesw5-dev xz-utils tk-dev libxml2-dev libxmlsec1-dev libffi-dev liblzma-dev
    return
  fi

  if is_pacman_distro; then
    pkg_install base-devel openssl zlib bzip2 readline sqlite curl git ncurses xz tk libxml2 libffi
    return
  fi

  return 1
}
