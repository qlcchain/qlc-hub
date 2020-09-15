# QLC-Hub [![Actions Status](https://github.com/qlcchain/qlc-hub/workflows/Main%20workflow/badge.svg)](https://github.com/qlcchain/qlc-hub/actions)

Cross-chain Hub between NEO and Ethereum...

## QLC-Hub CLI

```
qlc-hub 0.0.1-3cb92e8.2020-09-13T10:01:32Z+0800
Usage:
  ghub [OPTIONS]

Application Options:
  -V, --verbose              show verbose debug information
  -l, --level=               log level (default: warn)
      --signerToken=         singer JWT token
      --signerEndPoint=      singer endpoint
  -n, --neoUrl=              NEO RPC endpoint
      --neoContract=         NEO staking contract address
      --neoAssetId=          qlc token asset id
      --neoSignerAddress=    NEO address to sign tx
      --neoAssetsAddress=    NEO address to keep assets
      --neoConfirmedHeight=  Neo transaction Confirmed Height (default: 0)
      --neoDepositInterval=  Lock timeout interval height of deposit (default: 40)
      --neoWithdrawInterval= Lock timeout interval height of withdraw (default: 20)
  -e, --ethereumUrl=         Ethereum RPC endpoint
      --ethereumContract=    ethereum staking contract address
      --ethOwnerAddress=     Ethereum owner address
      --ethConfirmedHeight=  Eth transaction Confirmed Height (default: 0)
      --ethDepositHeight=    Lock timeout Height of deposit (default: 20)
      --ethWithdrawHeight=   Lock timeout Height of withdraw (default: 40)
      --gasEndPoint=         endpoint to get gas price
      --listenAddress=       RPC server listen address (default: tcp://0.0.0.0:19745)
      --grpcAddress=         GRPC server listen address (default: tcp://0.0.0.0:19746)
      --allowedOrigins=      AllowedOrigins of CORS (default: *)
      --minDepositAmount=    minimal amount to deposit (default: 100000000)
      --minWithdrawAmount=   minimal amount to withdraw (default: 100000000)
      --withdrawFrequency=   time interval to every withdraw (minute) (default: 10)
      --stateInterval=       time interval to check locker state (default: 2)
  -K, --key=                 private key
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