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
      }: let
        # Single source of truth for the version, bumped by release-please on
        # each release; goreleaser stamps the same tag into the release archives.
        version = (builtins.fromJSON (builtins.readFile ./.release-please-manifest.json)).".";

        # The same dependency pin the binary is built with. buildGoModule wires
        # GOPATH/GOMODCACHE to a fetched, read-only module cache, which lets the
        # static-analysis checks run in the Nix sandbox with no network.
        vendorHash = "sha256-5xdi1DKzb+gMgJgE2255Xpw1dcArzQEg+lPenDbyp+g=";

        # A checks.<system>.* derivation whose whole output is the gate result:
        # it reuses buildGoModule's offline module setup, runs one command, and
        # installs nothing.
        mkGoGate = name: script:
          pkgs.buildGoModule {
            pname = "m2d2-${name}";
            version = "0.0.0";
            src = ./.;
            inherit vendorHash;
            doCheck = false;
            nativeBuildInputs = [pkgs.golangci-lint pkgs.nilaway];
            buildPhase = ''
              runHook preBuild
              export HOME=$TMPDIR
              export GOCACHE=$TMPDIR/go-cache
              ${script}
              runHook postBuild
            '';
            installPhase = "touch $out";
            meta.description = "m2d2 ${name} gate";
          };
      in {
        packages.default = pkgs.buildGoModule {
          pname = "m2d2";
          inherit version vendorHash;
          src = ./.;
          subPackages = ["cmd/m2d2"];
          ldflags = ["-s" "-w" "-X main.version=${version}"];
          meta = {
            description = "Bidirectional converter between Mermaid and D2 diagram syntax";
            mainProgram = "m2d2";
          };
        };

        checks = {
          golangci-lint = mkGoGate "golangci-lint" "golangci-lint run ./...";
          nilaway = mkGoGate "nilaway" "nilaway -include-pkgs=github.com/noamsto/mermaid2d2 ./...";
          go-test-race = mkGoGate "go-test-race" "go test -race ./...";
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
          # Generated SVGs (mermaid/d2 renders) aren't prose to spell-check.
          typos.excludes = ["\\.svg$"];
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
              pkgs.nilaway
              config.treefmt.build.wrapper
            ];
        };
      };
    };
}
