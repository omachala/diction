{
  lib,
  buildGoModule,
  makeWrapper,
  ffmpeg,
}:
buildGoModule {
  pname = "diction-gateway";
  version = "6.1";

  src = lib.cleanSource ../.;
  modRoot = "gateway";
  subPackages = ["."];

  vendorHash = "sha256-SFXBQyziw/iZL4ss1kMnHphz2Hrfx2BnxdBpwnVUDyw=";

  ldflags = ["-s" "-w"];

  nativeBuildInputs = [makeWrapper];

  postInstall = ''
    mv $out/bin/gateway $out/bin/diction-gateway
    wrapProgram $out/bin/diction-gateway \
      --prefix PATH : ${lib.makeBinPath [ffmpeg]}
  '';

  meta = {
    description = "OpenAI-compatible STT gateway with WebSocket streaming for the Diction iOS keyboard";
    homepage = "https://github.com/omachala/diction";
    license = lib.licenses.mit;
    mainProgram = "diction-gateway";
    platforms = lib.platforms.linux ++ lib.platforms.darwin;
  };
}
