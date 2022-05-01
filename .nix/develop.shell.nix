{ pkgs, ... }:

with pkgs; mkShell {
  nativeBuildInputs = [ pkg-config ];
  buildInputs = [
    cargo
    clippy
    openssl
    rust-analyzer
    rustc
    rustfmt

    scdoc
  ];
}
