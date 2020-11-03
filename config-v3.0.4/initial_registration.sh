#!/bin/sh 

curl --insecure --location --request POST 'http://localhost:10000/registration/' \
--header 'Content-Type: application/json' \
--data-raw '{
    "authenticationMethod": "5G_AKA",
    "supiOrSuci": "2089300007487",
    "K": "5122250214c33e723a5dd523fc145fc0",
    "opcType": "OP",
	"opc": "c9e8763286b5b9ffbdf56e1297d0887b",
	"plmnId": "",
	"servingNetworkName": "",
    "n3IWFIpAddress": "n3iwf.my5Gcore.org",
    "SNssai": {
        "Sst": 1,
        "Sd": "010203"
    }
}'
