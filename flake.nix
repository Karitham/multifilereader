{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils/main";
  };
  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = import nixpkgs {inherit system;};
      in {
      devShell = pkgs.mkShell {
        name = "mfr";
        packages = with pkgs; [
          go_1_22
          gofumpt
        ];
      };
    });
}
