{ pkgs, pname, version, ... }:

let
  inherit (pkgs)
    lib
    openssl
    pkg-config
    rustPlatform
    scdoc;
in
rustPlatform.buildRustPackage {
  inherit pname;
  version = "v${version}";

  src = lib.cleanSource ../.;

  nativeBuildInputs = [ pkg-config ];
  buildInputs = [ openssl scdoc ];

  cargoSha256 = "sha256-4FrEvIA1xqV0TOqV+qiJye1/CJbyFTU6TUBZsLVE2gU=";

  # TODO: add this when we have documentation written with scdoc
  # makeFlags = [
  #   "PREFIX=$(out)"
  # ];

  # postBuild = ''
  #   make $makeFlags
  # '';

  # preInstall = ''
  #   make $makeFlags install
  # '';

  meta = with lib; {
    homepage = "https://sr.ht/~glorifiedgluer/mata";
    description = "A CLI tool for mataroa / mataroa.blog";
    license = licenses.mit;
    maintainers = with maintainers; [ ratsclub ];
  };
}
