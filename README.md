[![codecov](https://codecov.io/gh/dogechain-lab/dogechain/branch/dev/graph/badge.svg?token=EKFM6S9478)](https://codecov.io/gh/dogechain-lab/dogechain)
[![QA](https://github.com/dogechain-lab/dogechain/actions/workflows/actions.yml/badge.svg)](https://github.com/dogechain-lab/dogechain/actions/workflows/actions.yml)

## Important notification

Sorry about the delay. But there will not be any versions higher than v1.2.x anymore in this repo.

The Polygon-Edge based architecture does not seem to be able to afford the heavy I/O type transactions of the current Dogechain. And most users have suffered from network congestion, RPC endpoint crash, long time confirmation, and higher and higher node maintaining cost.

After several days discussion, We decided to switch it to a BSC based architecture for more stable use cases. This may take 4-6 months, but we still need to do it for the sake of all users.

Version will bump to v2.x, a HardFork is expected, and all data will be migrated. There should be some bumps, but in the end we will end up with more stable network, RPC endpoints, transaction maintaining/execution, smaller database size(after pruning), etc...


## Dogechain

Dogechain-Lab Dogechain is a modular and extensible framework for building Ethereum-compatible blockchain networks, which forks from 0xPolygon's Polygon Edge project.

To find out more about Polygon, read [this README](./README-PolygonEdge.md).

---

Copyright 2022 Dogechain-Lab

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
