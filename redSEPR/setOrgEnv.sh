#!/bin/bash
#
# SPDX-License-Identifier: Apache-2.0

# default to using CNE
ORG=${1:-CNE}

# Exit on first error, print all commands.
set -e
set -o pipefail

# Where am I?
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"

ORDERER_CA=${DIR}/redSEPR/organizations/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem
PEER0_CNE_CA=${DIR}/redSEPR/organizations/peerOrganizations/cne.example.com/tlsca/tlsca.cne.example.com-cert.pem
PEER0_MOE_CA=${DIR}/redSEPR/organizations/peerOrganizations/moe.example.com/tlsca/tlsca.moe.example.com-cert.pem
PEER0_ORG3_CA=${DIR}/redSEPR/organizations/peerOrganizations/org3.example.com/tlsca/tlsca.org3.example.com-cert.pem


if [[ ${ORG,,} == "cne" || ${ORG,,} == "digibank" ]]; then

   CORE_PEER_LOCALMSPID=CNEMSP
   CORE_PEER_MSPCONFIGPATH=${DIR}/redSEPR/organizations/peerOrganizations/cne.example.com/users/Admin@cne.example.com/msp
   CORE_PEER_ADDRESS=localhost:7051
   CORE_PEER_TLS_ROOTCERT_FILE=${DIR}/redSEPR/organizations/peerOrganizations/cne.example.com/tlsca/tlsca.cne.example.com-cert.pem

elif [[ ${ORG,,} == "moe" || ${ORG,,} == "magnetocorp" ]]; then

   CORE_PEER_LOCALMSPID=MOEMSP
   CORE_PEER_MSPCONFIGPATH=${DIR}/redSEPR/organizations/peerOrganizations/moe.example.com/users/Admin@moe.example.com/msp
   CORE_PEER_ADDRESS=localhost:9051
   CORE_PEER_TLS_ROOTCERT_FILE=${DIR}/redSEPR/organizations/peerOrganizations/moe.example.com/tlsca/tlsca.moe.example.com-cert.pem

else
   echo "Unknown \"$ORG\", please choose CNE/Digibank or MOE/Magnetocorp"
   echo "For example to get the environment variables to set upa MOE shell environment run:  ./setOrgEnv.sh MOE"
   echo
   echo "This can be automated to set them as well with:"
   echo
   echo 'export $(./setOrgEnv.sh MOE | xargs)'
   exit 1
fi

# output the variables that need to be set
echo "CORE_PEER_TLS_ENABLED=true"
echo "ORDERER_CA=${ORDERER_CA}"
echo "PEER0_CNE_CA=${PEER0_CNE_CA}"
echo "PEER0_MOE_CA=${PEER0_MOE_CA}"
echo "PEER0_ORG3_CA=${PEER0_ORG3_CA}"

echo "CORE_PEER_MSPCONFIGPATH=${CORE_PEER_MSPCONFIGPATH}"
echo "CORE_PEER_ADDRESS=${CORE_PEER_ADDRESS}"
echo "CORE_PEER_TLS_ROOTCERT_FILE=${CORE_PEER_TLS_ROOTCERT_FILE}"

echo "CORE_PEER_LOCALMSPID=${CORE_PEER_LOCALMSPID}"
