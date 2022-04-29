{ pkgs, pname, version, ... }:

let
  inherit (pkgs)
    buildGoModule
    lib
    pandoc;
in
buildGoModule {
  inherit pname;
  version = "v${version}";

  src = lib.cleanSource ../.;

  nativeBuildInputs = [ pandoc ];

  vendorSha256 = "sha256-Dw0bIIssEd8UEh+pnd+Nk7RP72+HvHTGwJlOpCOQRG8=";

  subPackages = [ "cmd/mata" ];

  makeFlags = [
    "PREFIX=$(out)"
  ];

  postBuild = ''
    make $makeFlags
  '';

  preInstall = ''
    make $makeFlags install
  '';

  meta = with lib; {
    homepage = "https://sr.ht/~glorifiedgluer/mata";
    description = "A CLI tool for mataroa / mataroa.blog";
    license = licenses.mit;
    maintainers = with maintainers; [ ratsclub ];
  };
}
