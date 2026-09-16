# Chromarium-Mechanicus

## Usage
```
go run main.go -t .config/template.json --templatesDir .config/ --themesDir themes test
```
```
go run main.go --help                                                                                via 🐹 v1.26.5 via 🌙 
usage: main [<flags>] <theme>


Flags:
      --[no-]help   Show context-sensitive help (also try --help-long and --help-man).
  -v, --[no-]debug  Enable debug mode.
  -c, --config="/home/christoph/.config/chromarium-mechanicus/config.json"
                    alternative config file to use instead of xdgConfigHome/chromarium-mechanicus/config.json
  -t, --template=/home/christoph/.config/chromarium-mechanicus/template.json
                    alternative template file to use instead of xdgConfigHome/chromarium-mechanicus/template.json
      --templatesDir="/home/christoph/.config/chromarium-mechanicus"
                    alternative directory to resolve relatives paths inside the template.json file to
      --themesDir="/home/christoph/.config/chromarium-mechanicus/themes"
                    alternative directory to search for themes in

Args:
  <theme>  Theme to use.
```

## Themes

A theme is any JSON file. There is no fixed schema: whatever keys the file
contains are exactly the keys a template can reference, nested to any depth.

```json
{
  "mode": "dark",
  "wallpaper": "/home/me/Pictures/nord.jpg",
  "font": { "family": "Iosevka", "size": 13 },
  "colors": {
    "background": "#2E3440",
    "primary": "#88C0D0",
    "palette": { "red": "#BF616A" }
  }
}
```

```
{{ .mode }}                  -> dark
{{ .font.family }}           -> Iosevka
{{ .colors.primary }}        -> #88c0d0
{{ .colors.palette.red }}    -> #bf616a
```

Keys that are not valid template identifiers (dashes, dots, spaces) are still
reachable with `index`:

```
{{ index .colors "primary-dim" }}
```

A key that does not exist is a hard error, not a silent empty string.

### Color formats

Every string that parses as a color gets the formats below. Input may be
`#RGB`, `#RGBA`, `#RRGGBB`, `#RRGGBBAA`, `rgb()`, `rgba()`, `hsl()` or
`hsla()`; strings that are not colors are passed through untouched.

| Name | Format | `{{ .colors.primary.<name> }}` for `#470228` |
| --- | --- | --- |
| *(none)* | `#RRGGBB` | `#470228` |
| `hex` | `#RRGGBB` | `#470228` |
| `hex_stripped` | `RRGGBB` | `470228` |
| `hex_alpha` | `#RRGGBBAA` | `#470228ff` |
| `hex_alpha_stripped` | `RRGGBBAA` | `470228ff` |
| `alpha_hex` | `#AARRGGBB` | `#ff470228` |
| `alpha_hex_stripped` | `AARRGGBB` | `ff470228` |
| `rgb` | `rgb(r, g, b)` | `rgb(71, 2, 40)` |
| `rgba` | `rgba(r, g, b, a)` | `rgba(71, 2, 40, 1.0)` |
| `hsl` | `hsl(h, s%, l%)` | `hsl(327.0, 94.5%, 14.3%)` |
| `hsla` | `hsla(h, s%, l%, a)` | `hsla(327.0, 94.5%, 14.3%, 1.0)` |
| `red` | 0 - 255 | `71` |
| `green` | 0 - 255 | `2` |
| `blue` | 0 - 255 | `40` |
| `alpha` | 0.0 - 1.0 | `1.0` |
| `hue` | 0.0 - 360.0 | `327.0` |
| `saturation` | 0.0 - 100.0 | `94.5%` |
| `lightness` | 0.0 - 100.0 | `14.3%` |

Written bare, a color renders as `hex`, so `{{ .colors.primary }}` and
`{{ .colors.primary.hex }}` are the same thing. Hex output is lowercase
regardless of how the theme file wrote it.

## Installation 

install nix or smth
```nix
inputs = {
  nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  chromarium-mechanicus = {
    url = "github:DerMaista/Chromarium-Mechanicus"; # https://github.com/DerMaista/Chromarium-Mechanicus
    inputs.nixpkgs.follows = "nixpkgs";
  };
};
```

### home-manager module

Import `chromarium-mechanicus.homeModules.default` into your home-manager
configuration:

```nix
{ inputs, ... }:
{
  imports = [ inputs.chromarium-mechanicus.homeModules.default ];

  programs.chromarium-mechanicus = {
    enable = true;

    wallpaperCmd = "swww img {{.wallpaper}}";

    templates = {
      # template from a file
      waybar = {
        source = ./templates/waybar.css;
        target = ".config/waybar/colors.css";
        postHook = "systemctl --user restart waybar";
      };

      # or inline
      kitty = {
        text = ''
          background {{.colors.background}}
          foreground {{.colors.primary}}
        '';
        target = ".config/kitty/theme.conf";
      };
    };

    themes = {
      nord = ./themes/nord.json;
      # or as a Nix attribute set
      gruvbox = {
        mode = "dark";
        wallpaper = "/home/me/Pictures/gruvbox.png";
        colors.primary = "#83a598";
      };
    };
  };
}
```

Then switch themes with:

```
chromarium-mechanicus nord
```

| Option | Description |
| --- | --- |
| `enable` | Install the package and generate the config files. |
| `package` | Package to use, defaults to this flake's. |
| `wallpaperCmd` | Command to set the wallpaper, `{{.wallpaper}}` is the image path. |
| `templates.<name>.source` | Template file. Mutually exclusive with `text`. |
| `templates.<name>.text` | Inline template content. Mutually exclusive with `source`. |
| `templates.<name>.target` | Destination path, or a list of them. Relative paths are resolved against the home directory. |
| `templates.<name>.preHook` / `.postHook` | Shell commands run before/after rendering, templated with the theme. |
| `themes.<name>` | Theme file or attribute set, installed as `<name>.json`. |
| `themesDir` | Look for themes elsewhere instead of the generated directory. |

`templates` becomes
`$XDG_CONFIG_HOME/chromarium-mechanicus/template.json`, `themesDir` becomes
`config.json`, and `themes` are written to
`$XDG_CONFIG_HOME/chromarium-mechanicus/themes/`.
