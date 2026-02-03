# 🛠️ Productivity & CLI Tools

This repository includes a curated selection of modern CLI tools designed to enhance productivity and improve the terminal experience for Fullstack Developers. These tools are automatically installed and configured across supported distributions (Arch/EndeavourOS and Ubuntu/Debian).

## 🚀 Included Tools

### Navigation & Shell

#### [`zoxide`](https://github.com/ajeetdsouza/zoxide) - A smarter cd command
**Command:** `z <directory>`
Zoxide remembers which directories you use most frequently. Instead of typing complex paths, you can jump to your favorite project folders from anywhere.
*   **Example:** `z project` instead of `cd ~/workspace/development/project`

#### [`starship`](https://starship.rs/) - The cross-shell prompt
**Context-aware prompt**
A minimal, fast, and customizable prompt that shows you exactly what you need (Git status, Node.js version, Python environment, Docker context) only when it's relevant.

### Git Workflow

#### [`lazygit`](https://github.com/jesseduffield/lazygit) - Simple terminal UI for git commands
**Command:** `lg` (alias) or `lazygit`
A powerful terminal user interface for Git. It allows you to commit, amend, rebase, and resolve conflicts visually without leaving the terminal.
*   **Features:** Interactive rebase, easy cherry-picking, line-by-line staging.

### Search & Discovery

#### [`fzf`](https://github.com/junegunn/fzf) - A general-purpose command-line fuzzy finder
**Shortcut:** `Ctrl + R` (History search)
Used extensively throughout the configuration to provide interactive filtering for command history, file searching, and more.

#### [`ripgrep`](https://github.com/BurntSushi/ripgrep) (rg) - Recursively search directories
**Command:** `rg "pattern"`
A line-oriented search tool that recursively searches the current directory for a regex pattern. It respects `.gitignore` rules automatically and is significantly faster than standard `grep`.

### Shortcuts & Aliases

#### `h` - Interactive History
**Command:** `h`
A supercharged history command. It pipes your command history into `fzf`, allowing you to fuzzy-search through past commands and execute them instantly.

### Utilities

#### [`bat`](https://github.com/sharkdp/bat) - A cat clone with wings
**Command:** `cat` (aliased)
A `cat` clone that supports syntax highlighting for a large number of programming languages and Git integration.

#### [`httpie`](https://httpie.io/) - Modern HTTP client
**Command:** `http`
A user-friendly command-line HTTP client for the API era. It comes with JSON support, colors, and formatted output out of the box.
*   **Example:** `http POST api.example.com/login username=admin`

#### [`tldr`](https://tldr.sh/) - Collaborative cheatsheets for console commands
**Command:** `tldr <command>`
Simplified and community-driven man pages. It gives you practical examples of the most common uses for a command.
*   **Example:** `tldr tar`

#### [`eza`](https://github.com/eza-community/eza) - A modern replacement for ls
**Command:** `ls` / `ll` (aliased)
A modern, maintained replacement for `ls` that uses colors to distinguish file types and metadata. It knows about symlinks, extended attributes, and git.

---

## 📦 Installation
All tools are defined in `.chezmoidata.yaml` and installed via the package manager appropriate for your distribution (`yay` for Arch/EndeavourOS, `apt`/scripts for Ubuntu) when you run `chezmoi apply`.
