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
```
inputs = {
  nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  chromarium-mechanicus = {
    url = "github:DerMaista/Chromarium-Mechanicus"; # https://github.com/DerMaista/Chromarium-Mechanicus
    inputs.nixpkgs.follows = "nixpkgs";
  };
};
```
