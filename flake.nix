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

          # WebView dependencies - usar webkitgtk 4.1 e criar compat layer
          webkitgtk_4_1
          gtk3
          glib
          glib-networking
          gsettings-desktop-schemas

          # Systray dependencies
          libayatana-appindicator
          libappindicator-gtk3
          libdbusmenu-gtk3

          # USB/HID
          libusb1

          # Build tools
          gcc
        ];

        # Variáveis necessárias para GTK/WebKit
        GIO_EXTRA_MODULES = "${pkgs.glib-networking}/lib/gio/modules";
        XDG_DATA_DIRS = "${pkgs.gsettings-desktop-schemas}/share/gsettings-schemas/${pkgs.gsettings-desktop-schemas.name}:$XDG_DATA_DIRS";

        shellHook = ''
          # Limpa e recria diretório de pkg-config
          rm -rf .pkgconfig
          mkdir -p .pkgconfig

          # Copia todos os .pc do webkitgtk e cria versões 4.0
          for pc in ${pkgs.webkitgtk_4_1.dev}/lib/pkgconfig/*.pc; do
            base=$(basename "$pc")
            cp "$pc" ".pkgconfig/$base"
            # Cria versão 4.0 se for 4.1, mas mantendo linkagem com 4.1
            newname=$(echo "$base" | sed 's/-4\.1/-4.0/g')
            if [ "$newname" != "$base" ]; then
              # Preserva as 5 primeiras linhas (definições de caminhos) e altera o resto
              head -n 5 "$pc" > ".pkgconfig/$newname"
              tail -n +6 "$pc" | sed -e "s/Name: WebKit2GTK/Name: WebKitGTK/g" \
                  -e "s/webkit2gtk-4\.1/webkit2gtk-4.0/g" \
                  -e "s/javascriptcoregtk-4\.1/javascriptcoregtk-4.0/g" \
                  -e "s/webkitgtk-4\.1/webkitgtk-4.1/g" \
                  >> ".pkgconfig/$newname"
              
              # Agora corrige especificamente as flags de Linkagem e Cflags para apontarem para 4.1
              sed -i -e "s/-lwebkit2gtk-4.0/-lwebkit2gtk-4.1/g" \
                     -e "s/-ljavascriptcoregtk-4.0/-ljavascriptcoregtk-4.1/g" \
                     -e "s/-I\''${includedir}\/webkitgtk-4.0/-I\''${includedir}\/webkitgtk-4.1/g" \
                     ".pkgconfig/$newname"
            fi
          done

          # Copia appindicator
          if [ -d ${pkgs.libayatana-appindicator}/lib/pkgconfig ]; then
            cp ${pkgs.libayatana-appindicator}/lib/pkgconfig/*.pc .pkgconfig/ 2>/dev/null || true
          fi

          export PKG_CONFIG_PATH="$(pwd)/.pkgconfig:$PKG_CONFIG_PATH"
          export PKG_CONFIG_PATH_FOR_TARGET="$(pwd)/.pkgconfig:$PKG_CONFIG_PATH_FOR_TARGET"
          export LD_LIBRARY_PATH="${pkgs.webkitgtk_4_1}/lib:${pkgs.gtk3}/lib:${pkgs.glib}/lib:$LD_LIBRARY_PATH"
          export CGO_ENABLED=1
          export CGO_CFLAGS="-I${pkgs.webkitgtk_4_1.dev}/include/webkitgtk-4.1"
          export CGO_LDFLAGS="-L${pkgs.webkitgtk_4_1}/lib -lwebkit2gtk-4.1"

          echo "ACC Jabra Agent - Desktop Mode"
          echo "WebKit e AppIndicator configurados"
        '';
      };
    };
}
