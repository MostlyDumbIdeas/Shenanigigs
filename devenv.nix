{ pkgs, lib, config, inputs, ... }:

{
  languages.go.enable = true;

  packages = with pkgs; [
    golangci-lint
    gopls
    delve
    nats-server
  ];

  services.redis.enable = true;

  pre-commit = {
    hooks = {
      golangci-lint = {
        enable = true;
        entry = "sh -c 'cd common && ${pkgs.golangci-lint}/bin/golangci-lint run && cd ../services/ingestion && ${pkgs.golangci-lint}/bin/golangci-lint run'";
        files = "\\.go$";
      };

      gofmt = {
        enable = true;
        entry = "gofmt -l -w";
        files = "\\.go$";
      };

      go-vet = {
        enable = true;
        entry = "sh -c 'cd common && go vet ./... && cd ../services/ingestion && go vet ./...'";
        files = "\\.go$";
      };

      go-mod-tidy = {
        enable = true;
        entry = "sh -c 'cd common && go mod tidy && cd ../services/ingestion && go mod tidy'";
        files = "go\\.mod$";
      };
    };
  };
}
