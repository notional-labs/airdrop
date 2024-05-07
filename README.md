# Traditional Cosmos Airdrop

You can use this software to take a snapshot of stakers and their bonded tokens at a given block height.  To verify that all accounts are represented, the total of the stakers bonded denom and the total staked supply are verified against each other as a final step.

Output is in JSON like:
```json
[
 {
  "address": "eve1nrzfre4u4mgxtz0p6jj5v2z3aa63jfcwpyhn6g",
  "coins": [
   {
    "denom": "eve",
    "amount": "626"
   }
  ]
 },
 {
  "address": "eve1uzaaqwrneh6cwzcmurz5qnu0qfm6mevvmlhs6w",
  "coins": [
   {
    "denom": "eve",
    "amount": "39"
   }
  ]
 },
 ...
]
```

## Usage
To run the airdrop tool, specify the path to the configuration file using the -c flag and optionally provide the desired block height using the -h flag (default is the "latest" height).

```bash
make build
LOG_LEVEL=debug ./airdrop -c configs/config.toml.example
```

This repo has SDK 47 dependencies and style, but you can change the version of the sdk referenced in go.mod and it should work on 50, 47, 46, and 45.


