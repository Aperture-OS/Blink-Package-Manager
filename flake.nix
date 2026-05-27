{
  description = "Blink Package Manager - Development Environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          config = {
            allowUnfree = false;
          };
        };

        blink = pkgs.buildGoModule {
          pname = "blink";
          version = "0.2.0-alpha";
          src = ./.;
          vendorHash = null;  # Use vendor directory or generate later
          nativeBuildInputs = with pkgs; [ gnumake ];
          meta = with pkgs.lib; {
            description = "Lightweight, source-based package manager for ApertureOS";
            homepage = "https://github.com/Aperture-OS/blink-package-manager";
            license = licenses.asl20;
            platforms = platforms.linux;
          };
        };

      in
      {
        devShells.default = pkgs.mkShell {
          name = "blink-dev-shell";

          inputsFrom = [ blink ];

          buildInputs = with pkgs; [
            go
            gopls
            gotools
            delve
            gnumake
            gcc
            git
            gnupg
            gnutar
            gzip
            bzip2
            xz
            unzip
            nil
            nixpkgs-fmt
            tree
            which
            shellcheck
            jq
          ];

          shellHook = ''
            export GOPATH="$PWD/.gopath"
            export GOBIN="$GOPATH/bin"
            export PATH="$GOBIN:$PATH"
            mkdir -p "$GOPATH"
            export BLINK_TEST_ROOT="$PWD/root"

            echo "╔══════════════════════════════════════════════════════════════╗"
            echo "║           Blink Package Manager - Flake Dev Shell            ║"
            echo "╠══════════════════════════════════════════════════════════════╣"
            echo "║  Go version: $(go version | awk '{print $3}')                "
            echo "║  GOPATH:     $GOPATH                                         "
            echo "║  Test root:  $BLINK_TEST_ROOT                                "
            echo "╚══════════════════════════════════════════════════════════════╝"
            echo ""
            echo "Flake inputs locked. Environment is reproducible."
            echo ""
            echo "Available commands:"
            echo "  make          - Build via Makefile"
            echo "  go build      - Build blink binary"
            echo "  go test ./... - Run tests"
            echo "  nix build     - Build blink package via Nix"
            echo ""
          '';

          GO111MODULE = "on";
          BLINK_GIT_PATH = "${pkgs.git}/bin/git";
          BLINK_GPG_PATH = "${pkgs.gnupg}/bin/gpg";
        };

        packages = {
          default = blink;
          blink = blink;
        };

        apps = {
          default = flake-utils.lib.mkApp {
            drv = blink;
          };
        };
      });
}
