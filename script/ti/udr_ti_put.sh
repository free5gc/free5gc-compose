set -e

curl -X PUT -H "Content-Type: application/json" --data @./udr_ti_data.json \
	http://udr.free5gc.org:8000/nudr-dr/v2/application-data/influenceData/1

exit 0
