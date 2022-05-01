{ pkgs, ... }:

with pkgs; mkShell {
  buildInputs = [
    cargo
    clippy
    rust-analyzer
    rustc
    rustfmt
  ];
}
