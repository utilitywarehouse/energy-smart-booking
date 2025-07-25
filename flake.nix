{
  description = "devshell for energy-smart-booking";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (system:
    let pkgs = import nixpkgs { inherit system; };
    in with pkgs; {
      devShells.default =
        mkShell { buildInputs = [ go_1_21 gopls golangci-lint gotestsum ]; };
    });
}
