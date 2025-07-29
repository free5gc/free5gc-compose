set -e

curl -X DELETE -H "Content-Type: application/json" \
	http://udr.free5gc.org:8000/nudr-dr/v2/application-data/influenceData/1

exit 0
