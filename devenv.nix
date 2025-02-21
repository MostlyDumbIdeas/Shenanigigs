{ pkgs, lib, config, inputs, ... }:

{
  languages.go.enable = true;

  packages = with pkgs; [
    golangci-lint
    gopls
    delve
  ];

  pre-commit = {
    hooks = {
      golangci-lint = {
        enable = true;
        entry = "${pkgs.golangci-lint}/bin/golangci-lint run";
        files = "\\.go$";
      };

      gofmt = {
        enable = true;
        entry = "gofmt -l -w";
        files = "\\.go$";
      };

      go-vet = {
        enable = true;
        entry = "go vet";
        files = "\\.go$";
      };

      go-mod-tidy = {
        enable = true;
        entry = "go mod tidy";
        files = "go\\.mod$";
      };
    };
  };
}
