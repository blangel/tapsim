module github.com/halseth/bitcoinscriptsim

go 1.19

require (
	github.com/btcsuite/btcd v0.23.4
	github.com/btcsuite/btcd/btcec/v2 v2.1.3
	github.com/urfave/cli/v2 v2.23.7
)

require (
	github.com/btcsuite/btcd/btcutil v1.1.0 // indirect
	github.com/btcsuite/btcd/chaincfg/chainhash v1.0.1 // indirect
	github.com/btcsuite/btclog v0.0.0-20170628155309-84c8d2346e9f // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/decred/dcrd/crypto/blake256 v1.0.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 // indirect
	golang.org/x/sys v0.0.0-20200814200057-3d37ad5750ed // indirect
)

replace github.com/btcsuite/btcd => github.com/halseth/btcd v0.0.0-20230509090317-4c4732b4c40f