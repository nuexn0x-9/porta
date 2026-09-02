#!/usr/bin/env sh
# PORTA Linux & macOS Uninstaller
set -e

printf "\033[36m[i] Uninstalling PORTA from your system...\033[0m\n"

# 1. Remove binary
for BIN in "/usr/local/bin/porta" "${HOME}/.local/bin/porta"; do
    if [ -f "$BIN" ]; then
        if [ -w "$(dirname "$BIN")" ]; then
            rm -f "$BIN"
        elif command -v sudo >/dev/null 2>&1; then
            sudo rm -f "$BIN"
        fi
        printf "\033[32m[✓] Removed %s\033[0m\n" "$BIN"
    fi
done

# 2. Ask to purge runtime directory
PORTA_DIR="${HOME}/.porta"
if [ -d "$PORTA_DIR" ]; then
    printf "Do you want to delete runtime data and logs in %s? [y/N]: " "$PORTA_DIR"
    read -r CONFIRM
    case "$CONFIRM" in
        [Yy]*)
            rm -rf "$PORTA_DIR"
            printf "\033[32m[✓] Removed %s\033[0m\n" "$PORTA_DIR"
            ;;
        *)
            printf "\033[36m[i] Retained %s\033[0m\n" "$PORTA_DIR"
            ;;
    esac
fi

printf "\033[32m[✓] PORTA has been uninstalled successfully.\033[0m\n"
