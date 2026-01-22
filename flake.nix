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
            # Cria versão 4.0 se for 4.1
            newname=$(echo "$base" | sed 's/-4\.1/-4.0/g')
            if [ "$newname" != "$base" ]; then
              sed 's/4\.1/4.0/g' "$pc" > ".pkgconfig/$newname"
            fi
          done

          # Copia appindicator
          if [ -d ${pkgs.libayatana-appindicator}/lib/pkgconfig ]; then
            cp ${pkgs.libayatana-appindicator}/lib/pkgconfig/*.pc .pkgconfig/ 2>/dev/null || true
          fi

          export PKG_CONFIG_PATH="$(pwd)/.pkgconfig:$PKG_CONFIG_PATH"
          export CGO_ENABLED=1
          export CGO_CFLAGS="-I${pkgs.webkitgtk_4_1.dev}/include/webkitgtk-4.1"
          export CGO_LDFLAGS="-L${pkgs.webkitgtk_4_1}/lib -lwebkit2gtk-4.1"

          echo "ACC Jabra Agent - Desktop Mode"
          echo "WebKit e AppIndicator configurados"
        '';
      };
    };
}
