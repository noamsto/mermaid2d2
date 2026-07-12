{
  description = "Bidirectional converter between Mermaid and D2 diagram syntax";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    git-hooks-nix = {
      url = "github:cachix/git-hooks.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = inputs @ {flake-parts, ...}:
    flake-parts.lib.mkFlake {inherit inputs;} {
      systems = ["x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin"];

      imports = [
        inputs.treefmt-nix.flakeModule
        inputs.git-hooks-nix.flakeModule
      ];

      perSystem = {
        pkgs,
        config,
        ...
      }: {
        packages.default = pkgs.buildGoModule {
          pname = "m2d2";
          version = "0.1.0";
          src = ./.;
          vendorHash = "sha256-aVgU4s7D4M+l2TDTa/3sWeHjFDD9QX68uBv4fJmAtqc=";
          subPackages = ["cmd/m2d2"];
          ldflags = ["-s" "-w"];
          meta = {
            description = "Bidirectional converter between Mermaid and D2 diagram syntax";
            mainProgram = "m2d2";
          };
        };

        treefmt = {
          projectRootFile = "flake.nix";
          programs = {
            alejandra.enable = true;
            gofmt.enable = true;
          };
        };

        pre-commit.settings.hooks = {
          statix.enable = true;
          deadnix.enable = true;
          alejandra.enable = true;
          typos.enable = true;
          check-merge-conflicts.enable = true;
          trim-trailing-whitespace.enable = true;
          gofmt.enable = true;
          govet.enable = true;
        };

        devShells.default = pkgs.mkShell {
          inherit (config.pre-commit) shellHook;
          packages =
            config.pre-commit.settings.enabledPackages
            ++ [
              pkgs.go
              pkgs.gopls
              pkgs.gotools
              pkgs.golangci-lint
              config.treefmt.build.wrapper
            ];
        };
      };
    };
}
