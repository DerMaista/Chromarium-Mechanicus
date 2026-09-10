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

    wallpaperCmd = "swww img {{.Wallpaper}}";

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
          background {{.Colors.Background}}
          foreground {{.Colors.Primary}}
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
| `wallpaperCmd` | Command to set the wallpaper, `{{.Wallpaper}}` is the image path. |
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
