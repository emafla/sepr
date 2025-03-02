# SEPR - Web3-based Tally Sheet Processing System
SEPR is a prototype of a system designed to process station-level tally sheets, in elections
based on paper voting. It is based on the Hyperledger Fabric permissioned blockchain and IPFS.

The paper tally sheets are scanned to produce a TIFF image file. The blockchain registers
all transactions that occur during the processing of the digital tally sheets, until the
election results are totalized.

The TIFF images are stored in a local IPFS Kubo node.

## Contents of the repository

```
sepr
├── actas                       // contains the app and smart contract
│   ├── apps-go                 // application
│   │   ├── TXs                 // stores the histories of tally sheet transactions
│   │   ├── go.mod
│   │   ├── go.sum
│   │   ├── tallySheets.go      // application
│   │   ├── tallySheets.json    // fixed values of tally sheets; used to create them
│   │   ├── tallySheetsCID      // stores the tally sheets image files copied from IPFS
│   │   └── tallySheetsTIF      // stores sample tally sheets images in TIFF format
│   └── chaincode-go            // smart contract
│       └── chaincode
│           └── smartcontract.go
└── redSEPR
    ├── compose                 // docker compose configuration files
    ├── configtx                // configuration of genesis block
    ├── network.sh              // bring up the nodes, network and to deploy the chaincode
    └── organizations           // crypto material and membership information of organizations
```

## Installation

The prototype runs in a Hyperledger Fabric, Docker Desktop, IPFS Kubo node, and WSL2 (Windows Subsystem for Linux version 2) environment. To run the prototype, follow these steps:

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

- Add the bin directory to PATH
```bash
export PATH=/path/to/HLF/bin;$PATH
```

3. Copy/clone the **actas** and **redSEPR** directories to the HLF directory

## Bring up the blockchain
The blockchain is configured with two member organizations and one orderer.

- Bring up the orderer and peer nodes, set up the Fabric network, and create the channel.

```bash
cd redSEPR
./network.sh up createChannel -c seprchannel -ca
```

- Deploy the chaincode (smart contract)
```bash
./network.sh deployCC -ccn seprcc -ccp ../actas/chaincode-go -ccl go -c seprchannel
```

## Application

The application to access the blockchain is in the **actas/apps-go** directory.

- Build the application
Go to the application directory

```bash
cd ../actas/apps-go
go build
```

- Create the directories **tallySheetsCID** and **TXs**.
The application stores tally sheets images copied from IPFS and history transactions of
tally sheets in those directories.

```bash
mkdir tallySheetsCID
mkdir TXs
```

- Start the IPFS Kubo daemon

- Start the application. We have some protobuf registration conflict warnings. For the time being, please, set the GOLANG_PROTOBUF_REGISTRATION_CONFLICT environment variable to ignore those warnings.
```bash
GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn ./tallySheets
```

## How to use the application

The blockchain uses the following data structure to represent digital tally sheets.

```
type TallySheet struct {
  UUID      [16]byte // Unique tally sheet ID
  Province  int      // Province ID
  Canton    int      // Canton ID
  Parish    int      // Parish ID
  Center    int      // Polling center ID
  PStation  int      // Polling station ID
  RegVoters int      // Registered voters
  Status    int      // Tally sheet status
  Cid       string   // IPFS CID of tally sheet
  Blank     [4]int   // Blank votes
  Spoiled   [4]int   // Spoiled votes
  Cand1     [4]int   // Candidate 1 votes
  Cand2     [4]int   // Candidate 2 votes
  Voters    [4]int   // Total voters counted
}
```
- The first seven fields identify the polling station and remain unchanged.
- The `Status` field tracks the tally sheet's states during its processing.
- The `Cid` field stores the IPFS CID of the scanned tally sheet.
- Vote data fields store results from ICR, two manual verifications, and the final validated count.

At startup, the application creates and initializes tally sheets in the blockchain using data from the **tallySheets.json** file. The initial state of the tally sheets is saved in the file
**tallySheets.csv** file.

The application offers a set of options to interact with the blockchain. The application simulates
real world tally sheet processing.

```
   1.- Register a tally sheet image (stores TIFF image in IPFS)
   2.- Invalidate a tally sheet
   3.- Register results of a tally sheet
   4.- Display a tally sheet's status
   5.- Display election results
   6.- Save a tally sheet' history in \"./TXs/UUID-history.csv\")
   7.- Save all tally sheets to \"tallySheets.csv\"
   8.- Display a transaction (given its TxId)
   9.- Exit
```

The tally sheets are accessed by their UUID. The **tallySheets.csv** file contains the UUIDs of each tally sheet.

1. First, tally sheets are scanned to produce a TIFF image file, which is stored in the Kubo IPFS.
The corresponding IPFS CID is stored in the *Cid* field. You have to provide the UUID of the tally
sheet and the name of the TIFF file. The **tallySheetsTIF** directory contains sample TIFF images.

2. Tally sheets can be invalidated by election authorities. In this case, a new paper tally sheet is prepared and scanned. You have to provide the tally sheet UUID.

3. A digitization procedure extracts the vote results from the image. You provide the tally sheet
UUID and the vote results.

4. This option displays the current state of a tally sheet. You provide the tally sheet UUID.

5. This option computes the total election results.

6.- Save a tally sheet's history in \"./TXs/UUID-history.csv\".

7. Saves the current state of all tally sheets in the file "tallySheets.csv".

8. Displays transaction information. You provide the transaction ID, that can be obtained from the
"UUID-history.csv" files.
