# QLC-Hub [![Actions Status](https://github.com/qlcchain/qlc-hub/workflows/Main%20workflow/badge.svg)](https://github.com/qlcchain/qlc-hub/actions)

Cross-chain Hub between NEO and Ethereum...

## QLC-Hub CLI

```
qlc-hub 0.0.1-3cb92e8.2020-09-13T10:01:32Z+0800
Usage:
  ghub [OPTIONS]

Application Options:
  -V, --verbose             show verbose debug information
  -l, --level=              log level (default: info)
      --signerToken=        singer JWT token
      --signerEndPoint=     singer endpoint
      --neoUrls=            NEO RPC endpoint
      --neoContract=        NEO staking contract address
      --neoOwnerAddress=    NEO address to sign tx
      --neoConfirmedHeight= Neo transaction Confirmed Height (default: 1)
  -e, --ethUrl=             Ethereum RPC endpoint
      --ethContract=        ethereum staking contract address
      --ethOwnerAddress=    Ethereum owner address
      --ethConfirmedHeight= Eth transaction Confirmed Height (default: 3)
      --listenAddress=      RPC server listen address (default: tcp://0.0.0.0:19745)
      --grpcAddress=        GRPC server listen address (default: tcp://0.0.0.0:19746)
      --allowedOrigins=     AllowedOrigins of CORS (default: *)
  -K, --key=                private key
      --duration=

Help Options:
  -h, --help                 Show this help message

```

## QLC-Signer CLI

```
signer 0.0.1-3cb92e8.2020-09-13T10:14:43Z+0800
Usage:
  signer [OPTIONS]

Application Options:
  -V, --verbose      show verbose debug information
  -K, --key=         private key for JWT manager
      --duration=    JWT token validity duration (default: 8760h0m0s)
  -l, --level=       log level (default: warn)
      --neoAccounts= NEO private keys
      --ethAccounts= ETH private keys
      --grpcAddress= GRPC server listen address (default: tcp://0.0.0.0:19747)

Help Options:
  -h, --help         Show this help message

```

## Links & Resources
* [Yellow Paper](https://github.com/qlcchain/YellowPaper)
* [QLC Website](https://qlcchain.org)
* [Reddit](https://www.reddit.com/r/QLCChain/)
* [Medium](https://medium.com/qlc-chain)
* [Twitter](https://twitter.com/QLCchain)
* [Telegram](https://t.me/qlinkmobile)