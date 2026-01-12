{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    glfw3
    pkg-config
    xorg.libX11.dev
    xorg.libXcursor.dev
    xorg.libXrandr.dev
    xorg.libXinerama.dev
    xorg.libXi.dev
    xorg.libXxf86vm.dev
    mesa
    libGL
    alsa-lib
  ];
  
  shellHook = ''
    export LD_LIBRARY_PATH="${pkgs.mesa.drivers}/lib:${pkgs.libGL}/lib:$LD_LIBRARY_PATH"
  '';
}