{
  description = "Chromarium Mechanicus flake";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
  };

  outputs = inputs: {
    packages = builtins.mapAttrs (system: pkgs: {
      chromarium-mechanicus = pkgs.buildGoModule {
        pname = "chromarium-mechanicus";
        version = "0.1.0";

        src = inputs.self;

        # Hash of the fetched Go module dependencies. Update whenever
        # go.mod/go.sum change.
        vendorHash = "sha256-FmrPMMFjjtMD6yuS9weP7EZraVL9OiW8WYBcCid3MJ8=";

        meta = {
          description = "Theme templating engine";
          mainProgram = "chromarium-mechanicus";
        };
      };

      default = inputs.self.packages.${system}.chromarium-mechanicus;
    }) inputs.nixpkgs.legacyPackages;
  };
}
