# Web3-based Tally Sheet Processing System
SEPR is a prototype of a system designed to process station-level election tally sheets.

The prototype runs in a Hyperledger Fabric, Docker Desktop, and WSL2 (Windows Subsystem for Linux version 2) environment. To run the prototype, follow these steps:

1. Install the prerequisites for WSL2.
Follow the instructions in [*Hyperledger Fabric Prerequisites for Windows*](https://hyperledger-fabric.readthedocs.io/en/latest/prereqs.html#windows)

2. Install the Hyperledger Fabric binaries and docker images

- Create a directory for SEPR and go to that directory (HLF, for example):
```bash
mkdir HLF
cd HLF
```

- Download the Hyperledger Fabric installation script
```bash
curl -sSLO https://raw.githubusercontent.com/hyperledger/fabric/main/scripts/install-fabric.sh && chmod +x install-fabric.sh
```

- Install the Hyperledger Fabric binaries and docker images
```bash
./install-fabric.sh --fabric-version 2.5.11 binary docker
```

- Add the bin dir to PATH
```bash
export PATH=/path/to/HLF/bin;$PATH
```

3. Copy/clone the **actas** and **redSEPR** directories to the HLF directory


4. Start the blockchain

- Start the network and create the channel

```bash
cd redSEPR
./network.sh up createChannel -c seprchannel -ca
```

- Deploy  the chaincode
```bash
./network.sh deployCC -ccn seprcc -ccp ../actas/chaincode-go -ccl go -c seprchannel
```

5. Application
- Start the application. We have some protobuf registration conflict warnings. For the time being, please, set the GOLANG_PROTOBUF_REGISTRATION_CONFLICT environment variable to ignore those warnings.
```bash
Cd ../actas/apps-go
GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn ./tallySheet
```
