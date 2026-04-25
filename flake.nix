{
  description = "Diction self-hosted gateway: OpenAI-compatible STT proxy with WebSocket streaming.";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = import nixpkgs {inherit system;};
    in {
      packages.diction-gateway = pkgs.callPackage ./nix/package.nix {};
      packages.default = self.packages.${system}.diction-gateway;

      devShells.default = pkgs.mkShell {
        packages = with pkgs; [
          go
          gopls
          gotools
          golangci-lint
          ffmpeg
        ];
      };

      checks.diction-gateway = self.packages.${system}.diction-gateway;
    })
    // {
      nixosModules.default = import ./nix/module.nix;
      nixosModules.diction-gateway = self.nixosModules.default;
      overlays.default = final: _prev: {
        diction-gateway = final.callPackage ./nix/package.nix {};
      };
    };
}
