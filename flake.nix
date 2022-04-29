{
  description = "Mataroa CLI";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, utils }:
    let
      pname = "mata";
      version = "0.3.1";
    in
    {
      overlays.default = final: prev: {
        mata = final.callPackage ./.nix/package.nix {
          inherit final pname version;
        };
      };
    } //
    utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };

        inherit (pkgs) callPackage;
      in
      rec {
        # `nix build`
        packages."${pname}" = callPackage ./.nix/package.nix {
          inherit pkgs pname version;
        };
        packages.default = packages."${pname}";

        # `nix run`
        apps."${pname}" = utils.lib.mkApp {
          drv = packages."${pname}";
        };
        apps.default = apps."${pname}";

        # `nix develop`
        devShells = {
          default = callPackage ./.nix/develop.shell.nix { inherit pkgs; };
          ci = callPackage ./.nix/ci.shell.nix { inherit pkgs; };
        };
      });
}
