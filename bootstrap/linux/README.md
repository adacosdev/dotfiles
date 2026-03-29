# Linux bootstrap layer

This directory is a structural placeholder for Linux-specific machine
provisioning.

Current `run_*` scripts still live at repository root because they are executed
directly by chezmoi. As the repo evolves, Linux bootstrap logic should be
documented and gradually extracted into helpers that this layer owns.

Examples that belong here conceptually:

- package installation
- runtime installation
- fonts
- editor extensions
- Linux service setup
