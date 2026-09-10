# `self` is this flake, applied when the module is imported. It must NOT also
# be taken as a module argument: home-manager setups commonly pass their own
# `self` through extraSpecialArgs, which would shadow this one.
self:
{ config, lib, pkgs, ... }:

let
  inherit (lib) mkEnableOption mkIf mkOption types;

  cfg = config.programs.chromarium-mechanicus;
  jsonFormat = pkgs.formats.json { };

  pruned = lib.filterAttrs (_: v: v != null);

  toList = v: if lib.isList v then v else [ v ];

  absTarget = p:
    if lib.hasPrefix "/" p || lib.hasPrefix "$" p then
      p
    else
      "${config.home.homeDirectory}/${p}";

  sourceOf = name: t:
    if t.source != null then
      toString t.source
    else
      toString (pkgs.writeText "chromarium-mechanicus-${name}" t.text);

  templateType = types.submodule {
    options = {
      source = mkOption {
        type = types.nullOr (types.either types.path types.str);
        default = null;
        example = lib.literalExpression "./templates/waybar.css";
        description = ''
          Path to the template file. Mutually exclusive with
          {option}`text`.
        '';
      };

      text = mkOption {
        type = types.nullOr types.lines;
        default = null;
        example = ''
          @define-color background {{.Colors.Background}};
          @define-color primary {{.Colors.Primary}};
        '';
        description = ''
          Template content written verbatim to the Nix store. Mutually
          exclusive with {option}`source`.
        '';
      };

      target = mkOption {
        type = types.either types.str (types.listOf types.str);
        example = ".config/waybar/colors.css";
        description = ''
          Where the rendered template is written. Relative paths are
          interpreted relative to the home directory.

          A list renders the same template to several destinations; the hooks
          then run once per destination.
        '';
      };

      preHook = mkOption {
        type = types.nullOr types.str;
        default = null;
        example = "mkdir -p ~/.cache/foo";
        description = ''
          Shell command run before the template is rendered. The theme is
          available for templating, e.g. `{{.Colors.Primary}}`.
        '';
      };

      postHook = mkOption {
        type = types.nullOr types.str;
        default = null;
        example = "systemctl --user restart waybar";
        description = ''
          Shell command run after the template is rendered. The theme is
          available for templating, e.g. `{{.Wallpaper}}`.
        '';
      };
    };
  };

  templateEntries = lib.concatLists (lib.mapAttrsToList
    (name: t:
      map
        (target: pruned {
          source_file = sourceOf name t;
          output_file = absTarget target;
          pre_hook = t.preHook;
          post_hook = t.postHook;
        })
        (toList t.target))
    cfg.templates);

  templateJson = pruned {
    wallpaper_cmd = cfg.wallpaperCmd;
    templates = templateEntries;
  };

  settings = pruned { themesDir = cfg.themesDir; };

  themeStateFile =
    if cfg.themeStateFile != null then
      cfg.themeStateFile
    else
      "${config.xdg.cacheHome}/chromarium-mechanicus/theme";
in
{
  options.programs.chromarium-mechanicus = {
    enable = mkEnableOption "Chromarium Mechanicus, a theme templating engine";

    package = mkOption {
      type = types.package;
      default = self.packages.${pkgs.stdenv.hostPlatform.system}.chromarium-mechanicus;
      defaultText = lib.literalExpression "chromarium-mechanicus.packages.\${system}.chromarium-mechanicus";
      description = "The chromarium-mechanicus package to use.";
    };

    wallpaperCmd = mkOption {
      type = types.nullOr types.str;
      default = null;
      example = "swww img {{.Wallpaper}}";
      description = ''
        Command used to set the wallpaper. `{{.Wallpaper}}` is replaced with
        the path to the theme's wallpaper image.
      '';
    };

    theme = mkOption {
      type = types.nullOr types.str;
      default = null;
      example = "nord";
      description = ''
        Theme applied on every home-manager activation, by name, as listed in
        {option}`themes` or found in {option}`themesDir`.

        This is the declarative default. Once {option}`themeStateFile` exists
        it wins, so switching themes at runtime survives an activation.
        Leave null to only ever apply what that file names.
      '';
    };

    themeStateFile = mkOption {
      # Not types.path: this names a file that exists at runtime, and a path
      # literal would be copied into the store at evaluation time.
      type = types.nullOr types.str;
      default = null;
      defaultText = lib.literalExpression ''"''${config.xdg.cacheHome}/chromarium-mechanicus/theme"'';
      example = "/home/me/.cache/chromarium-mechanicus/theme";
      description = ''
        File holding the name of the currently active theme, a single line
        such as `nord`. After every home-manager activation the theme named
        here is re-applied, so the templates are rendered against the config
        that was just linked.

        Nothing is run when the file does not exist, so a theme switcher owns
        this file: write the theme name to it, then run
        `chromarium-mechanicus <theme>`.
      '';
    };

    templates = mkOption {
      type = types.attrsOf templateType;
      default = { };
      example = lib.literalExpression ''
        {
          waybar = {
            source = ./templates/waybar.css;
            target = ".config/waybar/colors.css";
            postHook = "systemctl --user restart waybar";
          };
        }
      '';
      description = ''
        Templates rendered by Chromarium Mechanicus, written to
        {file}`$XDG_CONFIG_HOME/chromarium-mechanicus/template.json`.
      '';
    };

    themes = mkOption {
      type = types.attrsOf (types.either types.path jsonFormat.type);
      default = { };
      example = lib.literalExpression ''
        {
          nord = ./themes/nord.json;
          gruvbox = {
            mode = "dark";
            wallpaper = "/home/me/Pictures/gruvbox.png";
            colors.primary = "#83a598";
          };
        }
      '';
      description = ''
        Themes to install into
        {file}`$XDG_CONFIG_HOME/chromarium-mechanicus/themes`. Each attribute
        is either a path to an existing theme file or the theme itself as a
        Nix attribute set. The attribute name is the theme name passed to the
        binary.
      '';
    };

    themesDir = mkOption {
      type = types.nullOr types.str;
      default = null;
      example = "/home/me/dotfiles/themes";
      description = ''
        Directory searched for themes, written to
        {file}`$XDG_CONFIG_HOME/chromarium-mechanicus/config.json`. Leave
        null to use the default themes directory next to the config file,
        which is where {option}`themes` installs to.
      '';
    };
  };

  config = mkIf cfg.enable {
    assertions = lib.mapAttrsToList
      (name: t: {
        assertion = (t.source == null) != (t.text == null);
        message = ''
          programs.chromarium-mechanicus.templates.${name}: exactly one of
          'source' and 'text' must be set.
        '';
      })
      cfg.templates
    ++ [{
      assertion = cfg.themes == { } || cfg.themesDir == null;
      message = ''
        programs.chromarium-mechanicus: 'themes' installs into the default
        themes directory, which 'themesDir' overrides. Set only one of them.
      '';
    }];

    home.packages = [ cfg.package ];

    xdg.configFile = {
      "chromarium-mechanicus/template.json".source =
        jsonFormat.generate "chromarium-mechanicus-template.json" templateJson;
    }
    // lib.optionalAttrs (settings != { }) {
      "chromarium-mechanicus/config.json".source =
        jsonFormat.generate "chromarium-mechanicus-config.json" settings;
    }
    // lib.mapAttrs'
      (name: theme: lib.nameValuePair "chromarium-mechanicus/themes/${name}.json" {
        source =
          if lib.isPath theme || lib.isDerivation theme then
            theme
          else
            jsonFormat.generate "chromarium-mechanicus-theme-${name}.json" theme;
      })
      cfg.themes;

    # Re-apply the active theme after every activation, so a switch renders
    # the templates against whatever config.json/template.json were just
    # linked. No flags: the module installs everything at the locations the
    # binary already looks in, and passing --config a store path would break
    # the relative themesDir lookup main.go does against its parent directory.
    home.activation.chromarium-mechanicus =
      lib.hm.dag.entryAfter [ "linkGeneration" ] ''
        # A theme switched at runtime outranks the declarative default, so
        # that switching a theme survives the next activation.
        if [ -r ${lib.escapeShellArg themeStateFile} ]; then
          chromariumTheme=$(cat ${lib.escapeShellArg themeStateFile})
        else
          chromariumTheme=${lib.escapeShellArg (if cfg.theme != null then cfg.theme else "")}
        fi

        if [ -n "$chromariumTheme" ]; then
          run ${lib.getExe cfg.package} $VERBOSE_ARG "$chromariumTheme" \
            || echo "chromarium-mechanicus: failed to apply theme '$chromariumTheme'" >&2
        fi
      '';
  };
}
