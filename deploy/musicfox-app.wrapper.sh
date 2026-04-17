#!/bin/bash
# Wrapper script to launch musicfox.app binary directly.
# This avoids the symlink issue where NSBundle.mainBundle() returns the wrong path.
# The symlink issue causes missing app icon and blocked notifications.

# Resolve the real path of this script (follow symlinks)
REAL_PATH="$(readlink -f "$0")"
SCRIPT_DIR="$(dirname "$REAL_PATH")"

# This script is in Contents/MacOS/, same directory as the musicfox binary
APP_PATH="$SCRIPT_DIR/go-musicfox"

exec "$APP_PATH" "$@"
