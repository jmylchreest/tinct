#!/usr/bin/env python3
"""
Tinct Notify-Send Output Plugin

Sends desktop notifications with colour palette information using notify-send.

USAGE:

  Simple:
    tinct generate -i image -p wallpaper.jpg -o notify-send

  Complex (with custom settings):
    tinct generate -i image -p wallpaper.jpg -o notify-send \\
      --plugin-args 'notify-send={
        "urgency": "critical",
        "timeout": 5000,
        "title": "New Theme!",
        "message": "Custom colours extracted",
        "icon": "applications-graphics"
      }'

REQUIREMENTS:
  - notify-send (libnotify-bin)
  - Ubuntu/Debian: sudo apt-get install libnotify-bin
  - Arch: sudo pacman -S libnotify
  - Fedora: sudo dnf install libnotify

Author: Tinct Contributors
License: MIT
"""

import json
import subprocess
import sys
from typing import Dict, List, Any


PLUGIN_INFO = {
    "name": "notify-send",
    "type": "output",
    "version": "1.0.1",
    "description": "Send desktop notifications with colour palette information",
    "enabled": False,
    "author": "Tinct Contributors",
    "requires": ["notify-send"],
}


def show_plugin_info() -> None:
    """Output plugin information as JSON."""
    print(json.dumps(PLUGIN_INFO, indent=2))


def send_notification(
    palette: Dict[str, Any], plugin_args: Dict[str, Any], dry_run: bool = False
) -> None:
    """
    Send a desktop notification with palette information.

    Args:
        palette: The categorised colour palette data
        plugin_args: Custom plugin arguments
        dry_run: If True, show what would be done without actually sending
    """
    theme_type = palette.get("theme_type", 0)
    colors = palette.get("colours") or {}
    all_colors = palette.get("all_colours") or []

    # Map theme_type integer to string
    theme_map = {0: "auto", 1: "dark", 2: "light"}
    theme_str = theme_map.get(theme_type, "unknown").title()

    # Get notification settings from plugin_args with defaults
    urgency = plugin_args.get("urgency", "normal")
    timeout = plugin_args.get("timeout", 10000)
    title_prefix = plugin_args.get("title_prefix", "")
    show_hex = plugin_args.get("show_hex", True)
    show_count = plugin_args.get("show_count", True)
    app_name = plugin_args.get("app_name", "Tinct")
    icon = plugin_args.get("icon", "preferences-desktop-theme")

    # Build notification title
    if title_prefix:
        title = f"{title_prefix} Tinct Colour Palette ({theme_str} Theme)"
    else:
        title = f"Tinct Colour Palette ({theme_str} Theme)"
    if "title" in plugin_args:
        title = plugin_args["title"]

    # Build notification body with semantic colours
    body_lines = []

    # Add semantic colour roles
    role_order = [
        ("background", "Background"),
        ("backgroundMuted", "Background Muted"),
        ("foreground", "Foreground"),
        ("foregroundMuted", "Foreground Muted"),
        ("accent1", "Accent 1"),
        ("accent2", "Accent 2"),
        ("accent3", "Accent 3"),
        ("accent4", "Accent 4"),
        ("danger", "Danger"),
        ("warning", "Warning"),
        ("success", "Success"),
        ("info", "Info"),
        ("notification", "Notification"),
    ]

    for role_key, role_label in role_order:
        if role_key in colors:
            color_data = colors[role_key]
            hex_color = color_data.get("hex", "").upper()
            if hex_color and show_hex:
                body_lines.append(f"  {role_label}: {hex_color}")

    # Add total colour count
    if all_colors and show_count:
        body_lines.append(f"\nTotal colours extracted: {len(all_colors)}")

    # Add custom message if provided
    if "message" in plugin_args:
        body_lines.append(f"\n{plugin_args['message']}")

    body = "\n".join(body_lines)

    # Build notify-send command
    cmd = ["notify-send"]

    # Add app name if specified
    if app_name:
        cmd.extend(["-a", app_name])

    # Add icon if specified
    if icon:
        cmd.extend(["-i", icon])

    # Add urgency level
    cmd.extend(["-u", urgency])

    # Add timeout
    cmd.extend(["-t", str(timeout)])

    # Add title and body
    cmd.extend([title, body])

    # Handle dry-run mode
    if dry_run:
        print("=== DRY-RUN MODE ===")
        print("Would send notification with command:")
        print(f"  {' '.join(cmd)}")
        print(f"\nTitle: {title}")
        print(f"Body:\n{body}")
        print("\nSettings:")
        print(f"  Urgency: {urgency}")
        print(f"  Timeout: {timeout}ms")
        print(f"  App Name: {app_name}")
        print(f"  Icon: {icon}")
        print("===================")
        return

    # Send notification using notify-send
    try:
        result = subprocess.run(
            cmd,
            check=True,
            capture_output=True,
            text=True,
        )
        print(f"✓ Notification sent successfully")
        if plugin_args.get("verbose", False):
            print(f"  Command: {' '.join(cmd)}")
            print(f"  Title: {title}")
            print(f"  Icon: {icon}")
            print(
                f"  Colours shown: {len([line for line in body_lines if ':' in line])}"
            )
    except subprocess.CalledProcessError as e:
        print(f"✗ Error: Failed to send notification: {e}", file=sys.stderr)
        if e.stderr:
            print(f"  stderr: {e.stderr}", file=sys.stderr)
        sys.exit(1)
    except FileNotFoundError:
        print(
            "✗ Error: notify-send command not found. Please install libnotify.",
            file=sys.stderr,
        )
        print("  Ubuntu/Debian: sudo apt-get install libnotify-bin", file=sys.stderr)
        print("  Arch: sudo pacman -S libnotify", file=sys.stderr)
        print("  Fedora: sudo dnf install libnotify", file=sys.stderr)
        sys.exit(1)


def validate_palette(palette: Dict[str, Any]) -> bool:
    """
    Validate the palette structure.

    Args:
        palette: The palette data to validate

    Returns:
        True if valid, False otherwise
    """
    if not isinstance(palette, dict):
        print("Error: Invalid palette format - expected JSON object", file=sys.stderr)
        return False

    if "colours" not in palette and "all_colours" not in palette:
        print(
            "Error: Invalid palette - missing 'colours' or 'all_colours' field",
            file=sys.stderr,
        )
        return False

    return True


def main() -> None:
    """Main entry point for the plugin."""
    # Check if --plugin-info flag is provided
    if len(sys.argv) > 1 and sys.argv[1] == "--plugin-info":
        show_plugin_info()
        sys.exit(0)

    # Read palette JSON from stdin
    try:
        palette_data = json.load(sys.stdin)
    except json.JSONDecodeError as e:
        print(f"Error: Failed to parse JSON input: {e}", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"Error: Failed to read input: {e}", file=sys.stderr)
        sys.exit(1)

    # Validate palette
    if not validate_palette(palette_data):
        sys.exit(1)

    # Extract plugin args and dry-run flag
    plugin_args = palette_data.get("plugin_args", {})
    dry_run = palette_data.get("dry_run", False)

    # Show info about plugin execution
    if plugin_args.get("verbose", False) or dry_run:
        print(f"\n{'=' * 50}")
        print(f"Tinct Notify-Send Plugin v{PLUGIN_INFO['version']}")
        print(f"{'=' * 50}")
        print(f"Dry-run mode: {dry_run}")
        print(
            f"Plugin args: {json.dumps(plugin_args, indent=2) if plugin_args else 'none'}"
        )
        print(f"{'=' * 50}\n")

    # Send notification
    send_notification(palette_data, plugin_args, dry_run)


if __name__ == "__main__":
    main()
