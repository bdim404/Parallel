{
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs =
    { self, nixpkgs }:
    {
      packages = nixpkgs.lib.genAttrs
        [
          "x86_64-linux"
          "aarch64-linux"
          "x86_64-darwin"
          "aarch64-darwin"
        ]
        (
          system:
          let
            pkgs = nixpkgs.legacyPackages.${system};
          in
          {
            default = pkgs.buildGoModule {
              pname = "parallel-socks";
              version = "unstable-${self.shortRev or "dirty"}";
              src = ./.;
              vendorHash = "sha256-Z8V1a3uJdG/lj6AP4Xly01MQSq/yBnB2/TuERrrj0o0=";
              buildPhase = "go build -ldflags='-s -w' -o parallel-socks ./src";
              installPhase = "mkdir -p $out/bin && cp parallel-socks $out/bin/";
            };
          }
        );
    };
}

