# Git Shortcuts
alias gs="git status" # > View git status
alias ga="git add ." # > Add all changes to git
alias gc="git commit -m" # > Create a commit with message
alias gp="git push" # > Push changes to repository

# Node/Project Shortcuts
alias ni="npm install" # > Install node dependencies
alias nd="npm run dev" # > Start node development server
alias nb="npm run build" # > Build node project

# System Shortcuts
alias cls="clear" # > Clear terminal


# Fast Navigation
alias ..="cd .." # > Go up one directory level
alias ...="cd ../.." # > Go up two directory levels

# Docker (Time saver)
alias d="docker" # > Docker base command
alias dc="docker-compose" # > Docker-compose base command
alias dcu="docker-compose up -d" # > Start docker services in background
alias dcd="docker-compose down" # > Stop and remove docker services
alias dcl="docker-compose logs -f" # > View docker logs in real time

# Network Tools (Check occupied ports)
alias ports="sudo lsof -i -P -n | grep LISTEN" # > View listening ports

# Update all packages from pacman, yay and chezmoi.
alias update-all='sudo pacman -Syu && yay -Sua && chezmoi update' # > Update system and dotfiles

# Bat instead of cat
alias cat='bat --paging=never' # > Use bat to view files without paging

# Docker cleanup
alias docker-clean='docker system prune -a --volumes' # > Clean up docker resources

alias ls='eza --icons --group-directories-first' # > List files with eza (icons and directories first)
alias ll='eza -lh --icons --grid --group-directories-first' # > List files in detail with eza
alias tree='eza --tree --icons' # > Show directory tree with eza

alias ce='chezmoi edit' # > Edit chezmoi config
alias ca='chezmoi apply' # > Apply chezmoi changes
alias cdz='chezmoi cd' # > Change to chezmoi directory
alias cs='chezmoi status' # > View chezmoi status
alias capply='chezmoi apply && source ~/.zshrc' # > Apply chezmoi and reload terminal
