{
  description: "Ambiente de desenvolvimento Go para ACC Jabra Agent com UI Nativa";

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
          pkg-config
          webkitgtk
          gtk3
          libusb1
        ];

        shellHook = ''
          export PKG_CONFIG_PATH="${pkgs.webkitgtk.dev}/lib/pkgconfig:${pkgs.gtk3.dev}/lib/pkgconfig"
          echo "ACC Jabra Agent - Desktop Mode"
        '';
      };
    };
}