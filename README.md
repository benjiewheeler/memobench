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
- `prio_fee`: The priority fee in Lamports _(optional, if omitted, no priority fee will be used)_
- `node_retries`: The number of retries the RPC will rebroadcast the transaction

> [!IMPORTANT]
> The priority fee is in lamports not microlamports
