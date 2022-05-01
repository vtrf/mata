{ pkgs, ... }:

with pkgs; mkShell {
  buildInputs = [
    cargo
    clippy
    rustc
    rustfmt
  ];
}
