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
          nodejs_20
          pkg-config
          webkitgtk_4_1 # Versão moderna estável
          gtk3
          libusb1
        ];

        shellHook = ''
          # Cria um diretório temporário para links do pkg-config
          mkdir -p .pkgconfig
          cp ${pkgs.webkitgtk_4_1.dev}/lib/pkgconfig/webkit2gtk-4.1.pc .pkgconfig/webkit2gtk-4.0.pc
          # Ajusta o nome dentro do arquivo .pc se necessário (geralmente não é, o nome do arquivo é o que importa para o pkg-config)
          
          export PKG_CONFIG_PATH="$(pwd)/.pkgconfig:$PKG_CONFIG_PATH"
          echo "ACC Jabra Agent - Desktop Mode"
        '';
      };
    };
}
