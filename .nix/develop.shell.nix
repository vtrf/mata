{ pkgs, ... }:

with pkgs; mkShell {
  nativeBuildInputs = [ pkg-config openssl ];
  buildInputs = [
    cargo
    clippy
    rust-analyzer
    rustc
    rustfmt

    scdoc
  ];
}
