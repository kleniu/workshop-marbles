#!/bin/bash
#
# Copyright Robert Kleniewski
#
# SPDX-License-Identifier: Apache-2.0
#
# Exit on first error
set -e

DRP=`dirname $(realpath $0)`
cd $DRP

stime=$(date +%s)

test -d ~/.hfc-key-store/ || mkdir ~/.hfc-key-store/
cp creds/* ~/.hfc-key-store/.

export COMPOSE_PROJECT_NAME=nodesdk
docker-compose -f docker-compose.yml down
docker-compose -f docker-compose.yml up -d ca.example.com orderer.example.com peer0.org1.example.com couchdb

echo "Waiting 10s for startup ..."; sleep 10

docker exec -e "CORE_PEER_LOCALMSPID=Org1MSP" -e "CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/msp/users/Admin@org1.example.com/msp" peer0.org1.example.com peer channel create -o orderer.example.com:7050 -c mychannel -f /etc/hyperledger/configtx/channel.tx
docker exec -e "CORE_PEER_LOCALMSPID=Org1MSP" -e "CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/msp/users/Admin@org1.example.com/msp" peer0.org1.example.com peer channel join -b mychannel.block



#echo "Wainting 3s for channel creation and peer join ..."; sleep 3

#docker-compose -f ./docker-compose.yml up -d cli

#echo "##### INSTALL"
##docker exec -e "CORE_PEER_LOCALMSPID=Org1MSP" -e "CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp" cli peer chaincode install -n testcc -v 1.0 -p github.com/testcc
#docker exec cli peer chaincode install -n mytestcc -v 1.0 -p testcc

#echo "##### INSTANTIATE"
#docker exec -e "CORE_PEER_LOCALMSPID=Org1MSP" -e "CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp" cli peer chaincode instantiate -o orderer.example.com:7050 -C mychannel -n testcc -v 1.0 -c '{"Args":[""]}' -P "OR ('Org1MSP.member','Org2MSP.member')"
#docker exec cli peer chaincode instantiate -o orderer.example.com:7050 -C mychannel -n mytestcc -v 1.0 -c '{"Args":[""]}' -P "OR('Org1MSP.member')"




#echo "Wait 5s for instantiate completed ..."; sleep 5
#docker exec -e "CORE_PEER_LOCALMSPID=Org1MSP" -e "CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp" cli peer chaincode invoke -o orderer.example.com:7050 -C mychannel -n testcc -c '{"function":"initLedger","Args":[""]}'

printf "\nExecution time : $(($(date +%s) - stime)) secs ...\n\n"
