{ pkgs, ... }:

with pkgs; mkShell {
  nativeBuildInputs = [ pkg-config ];
  buildInputs = [
    cargo
    clippy
    rustc
    rustfmt

    openssl
  ];
}
