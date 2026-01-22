{
  description = "Ambiente de desenvolvimento Go para ACC Jabra Agent com UI Nativa";

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
          webkitgtk_4_1 # Versão moderna estável
          gtk3
          libusb1
        ];

        shellHook = ''
          # Força o pkg-config a encontrar o webkitgtk mesmo se o nome do pacote variar
          export PKG_CONFIG_PATH="${pkgs.webkitgtk_4_1.dev}/lib/pkgconfig:${pkgs.gtk3.dev}/lib/pkgconfig"
          # Mock para webkit2gtk-4.0 se necessário (webview_go pode pedir 4.0 ou 4.1)
          ln -sf ${pkgs.webkitgtk_4_1.dev}/lib/pkgconfig/webkit2gtk-4.1.pc ${pkgs.webkitgtk_4_1.dev}/lib/pkgconfig/webkit2gtk-4.0.pc || true
          echo "ACC Jabra Agent - Desktop Mode"
        '';
      };
    };
}
