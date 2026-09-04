{
  description = "Go development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go_1_26       # Go 1.26 toolchain
            gopls         # Language server
            gotools       # goimports, etc.
            delve         # Debugger (dlv)
            golangci-lint # Linter
            sqlite        # SQL Client
          ];

          shellHook = ''
            source .env
            echo "Go $(go version) — ready for use"
          '';
        };
      });
}
