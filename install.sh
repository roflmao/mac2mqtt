#!/usr/bin/env bash
# install.sh — installs mac2mqtt as a launchd service
#
# Usage:
#   sudo ./install.sh          # root mode  (LaunchDaemon, all features)
#        ./install.sh          # user mode  (LaunchAgent, no CPU temp / fan speed)
#   sudo ./install.sh --root   # force root mode
#        ./install.sh --user   # force user mode
#
# Run from the directory containing the mac2mqtt binary and mac2mqtt.yaml.

set -euo pipefail

LABEL="com.bessarabov.mac2mqtt"
PLIST_NAME="${LABEL}.plist"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# ── Helpers ───────────────────────────────────────────────────────────────────
info() { echo "[install] $*"; }
die()  { echo "[install] ERROR: $*" >&2; exit 1; }

# ── Mode detection ────────────────────────────────────────────────────────────
if [[ "${1:-}" == "--root" ]]; then
    MODE=root
elif [[ "${1:-}" == "--user" ]]; then
    MODE=user
elif [[ "$EUID" -eq 0 ]]; then
    MODE=root
else
    MODE=user
fi

# ── Paths ─────────────────────────────────────────────────────────────────────
if [[ "$MODE" == "root" ]]; then
    INSTALL_DIR="/usr/local/mac2mqtt"
    PLIST_DST="/Library/LaunchDaemons/${PLIST_NAME}"
    PLIST_TEMPLATE="${SCRIPT_DIR}/com.bessarabov.mac2mqtt.daemon.plist.template"
else
    INSTALL_DIR="${HOME}/mac2mqtt"
    PLIST_DST="${HOME}/Library/LaunchAgents/${PLIST_NAME}"
    PLIST_TEMPLATE="${SCRIPT_DIR}/com.bessarabov.mac2mqtt.agent.plist.template"
fi

# ── Preflight checks ──────────────────────────────────────────────────────────
if [[ "$MODE" == "root" && "$EUID" -ne 0 ]]; then
    die "Root mode requires elevated privileges. Run: sudo $0 [--root]"
fi

if [[ "$MODE" == "user" && "$EUID" -eq 0 ]]; then
    die "User mode must not run as root (LaunchAgent would install into root's home).
Run without sudo:  $0 --user"
fi

[[ -f "${SCRIPT_DIR}/mac2mqtt" ]] \
    || die "mac2mqtt binary not found in ${SCRIPT_DIR}"

[[ -f "$PLIST_TEMPLATE" ]] \
    || die "plist template not found: ${PLIST_TEMPLATE}"

if [[ ! -f "${SCRIPT_DIR}/mac2mqtt.yaml" ]]; then
    if [[ -f "${SCRIPT_DIR}/mac2mqtt.yaml.example" ]]; then
        die "mac2mqtt.yaml not found. Create it from the example first:
  cp \"${SCRIPT_DIR}/mac2mqtt.yaml.example\" \"${SCRIPT_DIR}/mac2mqtt.yaml\"
  # then edit it with your MQTT broker settings"
    else
        die "mac2mqtt.yaml not found in ${SCRIPT_DIR}"
    fi
fi

# ── Unload opposite mode (prevent concurrent duplicate services) ──────────────
if [[ "$MODE" == "root" ]]; then
    # Unload the LaunchAgent of the user who invoked sudo, if present.
    # $SUDO_USER / $SUDO_UID are set by sudo; skip if running as root directly.
    if [[ -n "${SUDO_USER:-}" && -n "${SUDO_UID:-}" ]]; then
        OPPOSITE_PLIST="/Users/${SUDO_USER}/Library/LaunchAgents/${PLIST_NAME}"
        if [[ -f "$OPPOSITE_PLIST" ]]; then
            info "Unloading and removing existing user-mode LaunchAgent for ${SUDO_USER}..."
            launchctl bootout "gui/${SUDO_UID}" "$OPPOSITE_PLIST" 2>/dev/null || true
            rm "$OPPOSITE_PLIST"
        fi
    fi
else
    # Cannot unload a system LaunchDaemon without root — abort and guide the user.
    OPPOSITE_PLIST="/Library/LaunchDaemons/${PLIST_NAME}"
    if [[ -f "$OPPOSITE_PLIST" ]]; then
        die "A root-mode LaunchDaemon is already installed. Remove it first, then re-run this script:
  sudo launchctl unload ${OPPOSITE_PLIST}
  sudo rm ${OPPOSITE_PLIST}"
    fi
fi

# ── Unload existing same-mode service ────────────────────────────────────────
if [[ -f "$PLIST_DST" ]]; then
    info "Unloading existing service..."
    launchctl unload "$PLIST_DST" 2>/dev/null || true
fi

# ── Install binary ────────────────────────────────────────────────────────────
info "Installing binary to ${INSTALL_DIR}..."
mkdir -p "$INSTALL_DIR"
cp "${SCRIPT_DIR}/mac2mqtt" "${INSTALL_DIR}/mac2mqtt"
chmod +x "${INSTALL_DIR}/mac2mqtt"
# Remove macOS quarantine flag from downloaded binaries (harmless if absent)
xattr -d com.apple.quarantine "${INSTALL_DIR}/mac2mqtt" 2>/dev/null || true

# ── Install config (never overwrite an existing file) ─────────────────────────
if [[ ! -f "${INSTALL_DIR}/mac2mqtt.yaml" ]]; then
    cp "${SCRIPT_DIR}/mac2mqtt.yaml" "${INSTALL_DIR}/mac2mqtt.yaml"
    info "Copied mac2mqtt.yaml"
else
    info "Keeping existing ${INSTALL_DIR}/mac2mqtt.yaml (not overwritten)"
fi

# ── Install plist ─────────────────────────────────────────────────────────────
if [[ "$MODE" == "root" ]]; then
    info "Installing LaunchDaemon plist..."
    cp "$PLIST_TEMPLATE" "$PLIST_DST"
    chown root:wheel "$PLIST_DST"
    chmod 644 "$PLIST_DST"
else
    info "Installing LaunchAgent plist..."
    mkdir -p "${HOME}/Library/LaunchAgents"
    # Substitute the USERNAME placeholder with the actual home directory path
    sed "s|/Users/USERNAME/|${HOME}/|g" "$PLIST_TEMPLATE" > "$PLIST_DST"
    chmod 644 "$PLIST_DST"
fi

# ── Load service ──────────────────────────────────────────────────────────────
info "Loading service..."
launchctl load "$PLIST_DST"

# ── Done ──────────────────────────────────────────────────────────────────────
echo
echo "mac2mqtt installed successfully (${MODE} mode)."
echo "  Binary:  ${INSTALL_DIR}/mac2mqtt"
echo "  Config:  ${INSTALL_DIR}/mac2mqtt.yaml"
echo "  Service: ${PLIST_DST}"
echo
if [[ "$MODE" == "root" ]]; then
    echo "To check status:  sudo launchctl list | grep mac2mqtt"
    echo "To restart:       sudo launchctl kickstart -k system/${LABEL}"
    echo "To view logs:     tail -f /tmp/mac2mqtt.err"
else
    echo "To check status:  launchctl list | grep mac2mqtt"
    echo "To restart:       launchctl kickstart -k gui/$(id -u)/${LABEL}"
    echo "To view logs:     tail -f /tmp/mac2mqtt.err"
fi
