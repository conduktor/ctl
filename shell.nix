# Nix dev shell for conduktor ctl project
# if nix installed you can smply run `nix-shell` to enter the dev shell
# if nix-direnv is installed, just add `use nix` in your .envrc file
{ pkgs ? import <nixpkgs> {} }:
let
  unstableTarball = builtins.fetchTarball https://github.com/NixOS/nixpkgs/archive/nixos-unstable.tar.gz;
  pkgs = import <nixpkgs> {};
  unstable = import unstableTarball {};

  shell = pkgs.mkShell {
    buildInputs = [
      pkgs.go
      pkgs.gopls
      unstable.golangci-lint
      pkgs.pre-commit
      pkgs.python313Packages.detect-secrets
     ];
  };
in shell