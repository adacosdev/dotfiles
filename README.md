# 🚀 adacosdev Dotfiles

![Linux](https://img.shields.io/badge/OS-Linux-blue?logo=linux)
![Chezmoi](https://img.shields.io/badge/Managed%20by-Chezmoi-black?logo=chezmoi)
![VSCode](https://img.shields.io/badge/Editor-VS%20Code-007ACC?logo=visual-studio-code)

Mi configuración personal y profesional automatizada con **Chezmoi**. Diseñada para un flujo de trabajo **Fullstack** y optimizada para **Ubuntu** y **EndeavourOS**.

## ✨ Características principales

- 🎨 **VS Code Adaptativo:** La barra de estado cambia de color según el entorno (Trabajo/Personal).
- 🖋️ **Tipografía:** Instalación automática de Nerd Fonts (*JetBrainsMono, Iosevka, FiraCode, Hack*).
- 🛠️ **Entornos:** Configuración lista para Docker, Node.js (`fnm`) y Python (`pyenv`).
- 🐚 **Zsh & Warp:** Alias optimizados y gestión de plugins.

## 📥 Instalación rápida

Si estás en una instalación limpia, solo necesitas tener `git` y `chezmoi` instalados. Luego ejecuta:

```bash
chezmoi init --apply [https://github.com/TU_USUARIO/dotfiles](https://github.com/TU_USUARIO/dotfiles)
```

## 📂 Estructura de Scripts
Los scripts se ejecutan automáticamente en orden gracias a los prefijos de Chezmoi:

| Script	                        | Función |
|----------------------------------|---------|
| run_once_00_install-docker.sh |	Instala Docker y gestiona permisos de grupo. |
| run_once_01_install-runtimes.sh | Configura fnm y pyenv con sus dependencias. |
| run_once_install-fonts.sh | Descarga y actualiza las fuentes en ~/.local/share/fonts. |
| run_once_install-extensions.sh | Sincroniza tus extensiones de VS Code. |

## 🔧 Configuración por Entorno
Este repo utiliza plantillas de Chezmoi. Para cambiar entre perfil personal o de trabajo, edita el archivo de configuración:

```bash
chezmoi edit-config
```
Y asegúrate de que la variable entorno esté definida:
```toml
[data]
  entorno = "home" # o "home"
```

Hecho con ❤️ por [adacosdev](https://github.com/adacosdev)