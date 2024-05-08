# memobench

Solana RPC benchmarking tool

This is a simple CLI tool that allows you to send transactions to an RPC endpoint and measure their success rate and confirmation time. The goal of this tool is to test the success rate of transactions landing on chain with any specified Solana RPC endpoint.

> [!NOTE]
> Make sure you have enough SOL in your test account to cover the transaction fees.

> [!CAUTION]
> This tool works by sending transactions to the blockchain, if configured with a high transaction count and/or high priority fees, it can lead to draining your test accounts. **USE WITH CAUTION!**

## Usage

- Download the latest release for your OS and architecture from [the releases page](https://github.com/benjiewheeler/memobench/releases).
- Execute the binary in a command prompt or terminal.
  - Upon first execution it will create a sample `config.json` file and exit
  - Edit the `config.json` file as needed
- Execute the binary again to start the benchmark

### Configuration

- `private_key`: The private key of the test account (in base58 format)
- `rpc_url`: The RPC endpoint to benchmark
- `ws_url`: The WS endpoint to listen for transactions _(optional, if omitted, the RPC URL will be used)_
- `send_rpc_url`: The RPC endpoint to send transactions _(optional, if omitted, the RPC URL will be used)_
- `rate_limit`: The rate limit (in requests per second)
- `tx_count`: The number of transactions to send
- `prio_fee`: The priority fee in Lamports per Compute Unit _(optional, if omitted, no priority fee will be used)_
- `node_retries`: The number of retries the RPC will rebroadcast the transaction

> [!IMPORTANT]
> The priority fee is in lamports not microlamports

## How does it work?

This tool works by sending a predefined number (`tx_count`) of unique transactions to the specified RPC (`send_rpc_url` or `rpc_url`). And count how many of them made it to the blockchain.

The transactions are sent all at once in parallel if possible, the tool will make sure to stay under the defined `rate_limit` to avoid getting 429 errors from the RPC.

The transactions sent are simple memo program transactions that each contain a unique memo in the form of `memobench: Test <number> [<id>]`.
The `<number>` part is used to ensure the memo is unique and by extension the transaction is unique, the `<id>` part is used to differentiate between individual tests.

## You like this tool ?

Buy me a coffee :coffee: _`CoffeeFpEteoCSPgHeoj98Sb6LCzoG36PGdRbYwqSvLd`_

_or hire me_ ;)

[![Discord Badge](https://img.shields.io/static/v1?message=Discord&label=benjie_wh&style=flat&logo=discord&color=7289da&logoColor=7289da)](https://discordapp.com/users/789556474002014219)
[![Telegram Badge](https://img.shields.io/static/v1?message=Telegram&label=benjie_wh&style=flat&logo=telegram&color=229ED9)](https://t.me/benjie_wh)
[![Protonmail Badge](https://img.shields.io/static/v1?message=Email&label=ProtonMail&style=flat&logo=protonmail&color=6d4aff&logoColor=white)](mailto:benjiewheeler@protonmail.com)
[![Github Badge](https://img.shields.io/static/v1?message=Github&label=benjiewheeler&style=flat&logo=github&color=171515)](https://github.com/benjiewheeler)
