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
sh -c "$(curl -fsLS get.chezmoi.io)" -- init --apply adacosdev
```

O si ya tienes `chezmoi`:

```bash
chezmoi init --apply adacosdev
```

## 📂 Estructura de Scripts
Los scripts se ejecutan automáticamente en orden gracias a los prefijos de Chezmoi:

| Script	                        | Función |
|----------------------------------|---------|
| run_once_00_install-docker.sh |	Instala Docker y gestiona permisos de grupo. |
| run_once_01_install-runtimes.sh | Configura fnm y pyenv con sus dependencias. |
| run_once_install-fonts.sh | Descarga y actualiza las fuentes en ~/.local/share/fonts. |
| run_once_install-extensions.sh | Sincroniza tus extensiones de VS Code. |

## 📂 Documentación y Escalabilidad

El repositorio incluye guías para facilitar su mantenimiento:

- [📈 Guía de Escalabilidad](docs/SCALABILITY.md): Explica cómo añadir paquetes, nuevas distribuciones y entender la estructura de datos en `.chezmoidata.yaml`.
- [🛠️ Herramientas de Productividad](docs/TOOLS.md): Descubre cómo usar `h` (historial mejorado), `zoxide`, `lazygit` y otras utilidades incluidas.

## 🔧 Configuración por Entorno
Este repo utiliza plantillas de Chezmoi. La primera vez que ejecutes `chezmoi init`, se te preguntará por tu email y tipo de entorno (`personal` o `adaion`).

Si necesitas cambiarlo más tarde, puedes usar:

```bash
chezmoi init
```
O editar directamente el archivo local:
```bash
chezmoi edit-config
```

Hecho con ❤️ por [adacosdev](https://github.com/adacosdev)
