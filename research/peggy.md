# Peggy

## Intro

Note: If we consider the ripple-consensus-protocol as having instant finality, according to the peggy spec, we should use IBC. Obviously, IBC isn't production ready, but it is an interesting observation.

"This design does not describe how tokens move into the Cosmos pegzone. It assumes that the pegzone itself establishes connectivity over IBC to a hub or other chains."

I think this just means that this spec is written for the hub, and we assume the hub uses IBC with the pegzone to move tokens.

This is a reference spec for Ethereum <-> Tendermint two-way peg.

## Design and Components

Five logical components
* Ethereum smart contract
* Witness
* Pegzone
* Signer
* Relayer

