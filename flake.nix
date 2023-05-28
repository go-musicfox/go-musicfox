{

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    devenv.url = "github:cachix/devenv";
    nix2container.url = "github:nlewo/nix2container";
    nix2container.inputs.nixpkgs.follows = "nixpkgs";
    mk-shell-bin.url = "github:rrbutani/nix-mk-shell-bin";
    mission-control.url = "github:Platonic-Systems/mission-control";
    flake-root.url = "github:srid/flake-root";
  };

  outputs = inputs@{ flake-parts, nixpkgs, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [
        inputs.devenv.flakeModule
        inputs.flake-root.flakeModule
        inputs.mission-control.flakeModule
        inputs.flake-parts.flakeModules.easyOverlay
      ];
      systems = nixpkgs.lib.systems.flakeExposed;
      perSystem = { config, self', inputs', pkgs, system, final, ... }: {
        mission-control.scripts = {
          run = {
            description = "Run app";
            exec = "go run ./cmd/musicfox.go";
            category = "Dev Tools";
          };
        };

        packages.default = pkgs.callPackage ./nix { };
        overlayAttrs = {
          inherit (config.packages) go-musicfox;
        };
        packages.joshuto = pkgs.callPackage ./nix { };

        devShells.default = pkgs.mkShell {
          inputsFrom = [
            config.flake-root.devShell
            config.mission-control.devShell
            self'.devShells.my-shell
          ];
          nativeBuildInputs = with pkgs; [
            pkg-config
          ];
          buildInputs = with pkgs;[
            alsa-lib
            flac
          ];
        };
        devenv.shells.my-shell = {
          languages.go = {
            enable = true;
          };
          packages = [
          ];
          enterShell = ''
            echo $'\e[1;32mWelcom to go-musicfox project~\e[0m'
          '';
        };
      };
    };
}
