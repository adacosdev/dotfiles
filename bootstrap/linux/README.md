# Linux bootstrap layer

This directory is a structural placeholder for Linux-specific machine
provisioning.

Current `run_*` scripts still live at repository root because they are executed
directly by chezmoi. As the repo evolves, Linux bootstrap logic should be
documented and gradually extracted into helpers that this layer owns.

Current helper sources live under `bootstrap/linux/helpers/` and are sourced by
the root `run_*` entrypoints via `{{ .chezmoi.sourceDir }}` so chezmoi keeps the
same execution model while Linux logic becomes reusable.

Examples that belong here conceptually:

- package installation
- runtime installation
- fonts
- editor extensions
- Linux service setup
