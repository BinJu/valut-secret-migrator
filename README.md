## a vault secrets migration tool which can migrate data across vaults no matter the network interconnectivity

- online mode.
  run the [`verify` command](https://github.com/BinJu/vault-secret-migrator/blob/master/cmd-verify/main.go). e.g.:
      ```bash
      go run main.go -source-addr="https://localhost:8201" -source-token="SOURCE-TOKEN" --target-  addr="https://localhost:8200" --target-token="TARGET-TOKEN" --dry-run=true --debug=true --root-path=/concourse/main --root-path=/concourse/shared
      ```

      if `dry-run=true`, the tool does not apply the change to the target vault, just show you the differences. 
      if your vaults are in kubernetes pods, you could try `kubectl port-forward`
- offline mode.
  1. run `go run export > FILE` on the original vault
  1. run `go run import FILE` to import the secrets.
  
