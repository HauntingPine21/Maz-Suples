{
  description = "Entorno de desarrollo de Maz-Suplementos";
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  outputs = { self, nixpkgs }: let
    systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
    forAll = nixpkgs.lib.genAttrs systems;
  in {
    devShells = forAll (system: let pkgs = import nixpkgs { inherit system; }; in {
      default = pkgs.mkShell { packages = with pkgs; [ go_1_24 gnumake git mysql80 ]; };
    });
  };
}
