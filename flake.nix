{
  description = "A simple tool to show the memory usage of a process and its children";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.vmrss = pkgs.buildGoModule {
          pname = "vmrss";
          version = "0.1.0";
          src = ./.;
          vendorHash = null;
        };

        apps.default = {
          type = "app";
          program = "${self.packages.${system}.vmrss}/bin/vmrss";
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [ go ];
        };
      }
    );
}
