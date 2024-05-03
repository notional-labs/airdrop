# Traditional Cosmos Airdrop

You can use this software to take a snapshot of stakers and their bonded tokens at a given block height.  To verify that all accounts are represented, the total of the stakers bonded denom and the total staked supply are verified against each other as a final step.

Output is in JSON like account:3345t345345ucoin


## Usage

```bash
go run main.go 6969 --node "https://r-composable--apikey.gw.notionalapi.net:443"
```

... or your favorite variation of the above.

This repo has SDK 47 dependencies and style, but you can change the version of the sdk referenced in go.mod and it should work on 50, 47, 46, and 45.


