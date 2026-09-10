{
  description = "Chromarium Mechanicus flake";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";

    # Only used by the test suite. Point this at your own home-manager with
    # inputs.chromarium-mechanicus.inputs.home-manager.follows = "home-manager";
    home-manager = {
      url = "github:nix-community/home-manager";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = inputs:
  let
    # Systems the test suite is evaluated for.
    testSystems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
  in
  {
    packages = builtins.mapAttrs (system: pkgs: {
      chromarium-mechanicus = pkgs.buildGoModule {
        pname = "chromarium-mechanicus";
        version = "0.1.0";

        src = inputs.self;

        vendorHash = "sha256-FmrPMMFjjtMD6yuS9weP7EZraVL9OiW8WYBcCid3MJ8=";

        meta = {
          description = "Theme templating engine";
          mainProgram = "chromarium-mechanicus";
        };
      };

      default = inputs.self.packages.${system}.chromarium-mechanicus;
    }) inputs.nixpkgs.legacyPackages;

    homeModules = {
      chromarium-mechanicus = import ./nix/hm-module.nix inputs.self;
      default = inputs.self.homeModules.chromarium-mechanicus;
    };

    # Deprecated alias, home-manager renamed this output to homeModules.
    homeManagerModules = inputs.self.homeModules;
  };
}
