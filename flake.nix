{
  description: "Ambiente de desenvolvimento Go para ACC Jabra Agent";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
    in {
      devShells.${system}.default = pkgs.mkShell {
        buildInputs = with pkgs; [
          go
          gopls
          go-tools
          pkg-config
          libusb1
        ];

        shellHook = ''
          echo "ACC Jabra Agent Development Environment"
          echo "Go: $(go version)"
        '';
      };
    };
}
