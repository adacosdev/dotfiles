#!/bin/bash

# Shared package-install helpers for Linux bootstrap entrypoints.

if ! command -v pkg_install >/dev/null 2>&1; then
  source "$(dirname "${BASH_SOURCE[0]}")/package-managers.sh"
fi

ensure_local_bin_symlink() {
  local source_path="$1"
  local target_path="$2"

  mkdir -p "$(dirname "$target_path")"
  ln -sf "$source_path" "$target_path"
}

install_starship_if_missing() {
  if command -v starship &> /dev/null; then
    return
  fi

  update_progress 50 "Installing starship prompt..."
  log_activity "Downloading and installing starship"
  curl -sS https://starship.rs/install.sh | sh -s -- -y 2>&1 | grep -i "installed\|done" || true
}

install_eza_if_missing() {
  if command -v eza &> /dev/null; then
    return
  fi

  update_progress 55 "Installing eza (modern ls)..."
  log_activity "Adding eza repository and installing"
  sudo mkdir -p /etc/apt/keyrings
  wget -qO- https://raw.githubusercontent.com/eza-community/eza/main/deb.asc | sudo gpg --dearmor -o /etc/apt/keyrings/gierens.gpg
  echo "deb [signed-by=/etc/apt/keyrings/gierens.gpg] http://deb.gierens.de stable main" | sudo tee /etc/apt/sources.list.d/gierens.list
  sudo chmod 644 /etc/apt/keyrings/gierens.gpg /etc/apt/sources.list.d/gierens.list
  pkg_update
  pkg_install eza 2>&1 | grep -i "setting up\|done" || true
}

ensure_bat_alias_if_available() {
  if command -v batcat &> /dev/null && ! command -v bat &> /dev/null; then
    update_progress 60 "Creating bat symlink..."
    log_activity "Setting up: ln -s /usr/bin/batcat ~/.local/bin/bat"
    ensure_local_bin_symlink /usr/bin/batcat "$HOME/.local/bin/bat"
  fi
}

install_lazygit_if_missing() {
  local lazygit_version

  if command -v lazygit &> /dev/null; then
    return
  fi

  update_progress 65 "Installing lazygit..."
  log_activity "Fetching latest lazygit release"
  lazygit_version=$(curl -s "https://api.github.com/repos/jesseduffield/lazygit/releases/latest" | grep -Po '"tag_name": "v\K[^"]*')
  curl -Lo /tmp/lazygit.tar.gz "https://github.com/jesseduffield/lazygit/releases/latest/download/lazygit_${lazygit_version}_Linux_x86_64.tar.gz"
  log_activity "Installing lazygit v${lazygit_version}"
  tar xf /tmp/lazygit.tar.gz -C /tmp lazygit
  sudo install /tmp/lazygit /usr/local/bin
  rm /tmp/lazygit /tmp/lazygit.tar.gz
}

install_delta_if_missing() {
  local delta_version

  if command -v delta &> /dev/null; then
    return
  fi

  update_progress 70 "Installing git-delta..."
  log_activity "Fetching latest git-delta release"
  delta_version=$(curl -s "https://api.github.com/repos/dandavison/delta/releases/latest" | grep -Po '"tag_name": "\K[^"]*')
  curl -Lo /tmp/git-delta.deb "https://github.com/dandavison/delta/releases/download/${delta_version}/git-delta_${delta_version}_amd64.deb"
  log_activity "Installing git-delta ${delta_version}"
  sudo dpkg -i /tmp/git-delta.deb
  rm /tmp/git-delta.deb
}

ensure_fd_alias_if_available() {
  if ! command -v fdfind &> /dev/null; then
    return
  fi

  if ! command -v fd &> /dev/null; then
    update_progress 75 "Creating fd symlink..."
    log_activity "Setting up: ln -s /usr/bin/fdfind ~/.local/bin/fd"
  fi

  ensure_local_bin_symlink /usr/bin/fdfind "$HOME/.local/bin/fd"
}
