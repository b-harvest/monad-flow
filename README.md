# monad-flow
Debugging tool to trace, inspect, and visualize the data flow (packets) inside your monad

```bash
jq -s 'map({(.secp): {name: .name, logo: .logo, description: .description, website: .website}}) | add' ./validator-info/testnet/*.json > validators.json
```
