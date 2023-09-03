{ lib
, buildGoModule
, pkg-config
, alsa-lib
, flac
, version ? "git"
}:

buildGoModule rec {
  pname = "go-musicfox";
  inherit version;

  src = ../../.;

  modRoot = ../../.;

  vendorHash = null;

  deleteVendor = true;

  subPackages = [ "cmd/musicfox.go" ];

  ldflags = [
    "-s"
    "-w"
    "-X github.com/go-musicfox/go-musicfox/internal/constants.AppVersion=${version}"
  ];

  nativeBuildInputs = [
    pkg-config
  ];

  buildInputs = [
    alsa-lib
    flac
  ];

  meta = with lib; {
    description = "Terminal netease cloud music client written in Go";
    homepage = "https://github.com/go-musicfox/go-musicfox";
    license = licenses.mit;
    mainProgram = "musicfox";
  };
}
