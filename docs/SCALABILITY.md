# 📈 Escalabilidad y Mantenimiento

Este repositorio está diseñado para ser escalable y soportar múltiples distribuciones de Linux (actualmente **Arch/EndeavourOS** y **Ubuntu/Debian**). Aquí se explica cómo añadir nuevas configuraciones o paquetes.

## 📦 Gestión de Paquetes

Los paquetes no se definen en los scripts, sino en `.chezmoidata.yaml`. Esto centraliza la gestión y facilita la lectura.

### Estructura de `.chezmoidata.yaml`

```yaml
packages:
  # Lista para Ubuntu/Debian
  ubuntu:
    - packagename
    - another-package
  # Lista para Arch/EndeavourOS
  arch:
    - packagename
    - another-package-aur
```

### Cómo añadir un paquete

1. Abre `.chezmoidata.yaml`.
2. Busca la sección de tu distribución (`arch` o `ubuntu`).
3. Añade el nombre exacto del paquete.
   - En **Arch**, el script usará `yay` (soporta repos oficiales y AUR).
   - En **Ubuntu**, el script usará `apt`.

### Cómo añadir una nueva Distribución (Fedora, por ejemplo)

1. Añade una nueva lista en `.chezmoidata.yaml`:
   ```yaml
   packages:
     fedora:
       - git
       - zsh
   ```
2. Edita `run_onchange_after_install-packages.sh.tmpl`:
   - Añade un bloque condicional:
     ```bash
     {{ else if eq .chezmoi.osRelease.id "fedora" -}}
     packages=(
     {{ range .packages.fedora -}}
       {{ . }}
     {{ end -}}
     )
     sudo dnf install -y "${packages[@]}"
     ```

## 🖥️ Configuración de Scripts

Los scripts utilizan plantillas de `chezmoi` (`.tmpl`). Puedes usar lógica de Go templates para condicionar la ejecución.

### Variables Útiles

- `{{ .chezmoi.os }}`: `linux`, `darwin` (macOS), `windows`.
- `{{ .chezmoi.osRelease.id }}`: `arch`, `ubuntu`, `debian`, `fedora`.
- `{{ .entorno }}`: Variable personalizada definida en el init (`personal` o `adaion`).

Ejemplo de uso en un script:
```bash
{{ if eq .entorno "adaion" }}
# Configuración específica del trabajo
{{ end }}
```

## 📂 Estructura de Archivos Recomendada

- **Scripts de instalación única:** `run_once_*.sh.tmpl` (se ejecutan solo si no existen o cambian).
- **Scripts de cambio:** `run_onchange_*.sh.tmpl` (se ejecutan cada vez que cambias el contenido del script, útil para listas de paquetes).
- **Configuraciones:** Usa `dot_config/carpeta/archivo` para mapear a `~/.config/carpeta/archivo`.

---

Mantén este archivo actualizado si cambias la lógica principal de instalación.
