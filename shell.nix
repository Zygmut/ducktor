{ pkgs? import (fetchTarball "https://github.com/NixOS/nixpkgs/tarball/nixos-24.05") {} }:

pkgs.mkShellNoCC {
  packages = with pkgs; [
    # https://pre-commit.com/
    pre-commit
    golangci-lint
    gotools

    # https://taskfile.dev/usage/
    go-task

    go
  ];

  shellHook = ''
    alias t=task

    pre-commit install

    task init
  '';

}
