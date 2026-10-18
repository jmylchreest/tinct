---
title: rosec
sidebar_position: 1
plugin:
  type: output
  name: rosec
  category: special
  source: builtin
  app: ROsec
  app_url: 'https://github.com/jmylchreest/rosec'
  requires: []
  optional: []
  pattern: single-file
  default_output_dir: ~/.config/rosec
  generated_files:
    - tinct-theme.toml
  reload:
    method: none
    user_action_required: true
---

# rosec

Generates a [ROsec](https://github.com/jmylchreest/rosec) prompt theme — the colours used by the `rosec-prompt` interactive prompt for backgrounds, borders, labels, inputs, and the success/danger button pair.

## Installation

Built into tinct — nothing to install separately. `tinct generate -o rosec` works out of the box.

The plugin skips silently if neither `~/.config/rosec/config.toml` exists nor `rosec-prompt` is on `$PATH`, so it's safe to leave in a multi-output invocation on hosts where rosec isn't installed.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o rosec
```

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct-theme.toml` | `~/.config/rosec/tinct-theme.toml` | TOML fragment containing a single `[prompt.theme]` table with every colour key rosec accepts. Not a complete rosec config — it's a partial that you merge into your `config.toml`. |

## Integration

rosec does not have an `include` or `import` mechanism, so you have to merge `[prompt.theme]` into your own `config.toml` by hand or with a TOML merge tool of your choice. The simplest copy-paste approach:

```bash
# View what tinct generated
cat ~/.config/rosec/tinct-theme.toml

# Manually replace the [prompt.theme] section in your config:
$EDITOR ~/.config/rosec/config.toml
```

If you maintain your rosec config under version control, a common pattern is to keep your non-theme settings in `config.toml` and use a small wrapper script that concatenates `config.toml` with `tinct-theme.toml` into rosec's effective config on demand.

## Reload behaviour

### Automatic

None — `Hooks().Reload` is nil. rosec reads its config on each prompt invocation, so any new prompt after you've merged the theme will already use the new colours; there is nothing to signal.

### Manual fallback

After merging `tinct-theme.toml` into `config.toml`, just trigger a new prompt — colours update on the next `rosec-prompt` invocation.

## Uninstall / revert

1. **Remove the config block** from `~/.config/rosec/config.toml`: delete (or revert) the `[prompt.theme]` table you merged in from `tinct-theme.toml`.

2. **Delete the generated file**:

   ```bash
   rm ~/.config/rosec/tinct-theme.toml
   ```

3. **Reload to drop the theme**: nothing to reload — the next `rosec-prompt` invocation will use whatever `[prompt.theme]` remains in `config.toml` (or rosec's defaults if you removed it entirely).

4. **External state**: this plugin only writes to `~/.config/rosec/`. No further cleanup is required.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--rosec.output-dir` | `~/.config/rosec` | Override the output directory |

## Colour role mapping

| `[prompt.theme]` key | Tinct role | Notes |
|---|---|---|
| `background` | `background` (+ `cc` alpha) | Prompt background, ~80% opacity |
| `foreground` | `foreground` | Main text |
| `border_color` | `accent1` | Border around the prompt |
| `label_color` | `foregroundMuted` | Field labels |
| `accent_color` | `accent1` | Highlighted UI bits |
| `danger_color` | `danger` | Destructive states |
| `input_background` | `surface` | Text-input field |
| `input_text` | `foreground` | Text inside the input field |
| `confirm_background` | `success` | Confirm button background |
| `confirm_text` | `foreground` | Confirm button text |
| `cancel_background` | `backgroundMuted` | Cancel button background |
| `cancel_text` | `foregroundMuted` | Cancel button text |

All colours except `background` are emitted with full alpha via `hexAlpha`; `background` uses an explicit `cc` suffix (~80% opacity).

## Customising the template

Dump the default template to override it:

```bash
tinct plugins templates dump -o rosec -l ~/.config/tinct/templates/rosec
```

This writes `~/.config/tinct/templates/rosec/tinct-theme.toml.tmpl`. Tinct uses your version in preference to the embedded default.

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for available functions.

## Troubleshooting

### "Plugin skipped: neither rosec config nor rosec-prompt binary found"

The plugin only writes when there's something to write to. Install `rosec-prompt`, or create `~/.config/rosec/config.toml` (even empty) to opt in.

### Theme generates but rosec still looks the same

`tinct-theme.toml` is a fragment, not a full config. Until you merge `[prompt.theme]` into `config.toml`, rosec doesn't see it. See [Integration](#integration).

### Background looks too transparent / too opaque

The template hard-codes the background alpha at `cc` (~80%). Dump the template and change the suffix on the `background` line (or pipe through `hexAlpha` instead).

## Related plugins

- [dunst](../bars-launchers/dunst.md), [histui](./histui.md) — notification daemons themed alongside the prompt.
