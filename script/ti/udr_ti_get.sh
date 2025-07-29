set -e

curl -X GET -H "Content-Type: application/json" \
	http://udr.free5gc.org:8000/nudr-dr/v2/application-data/influenceData/
echo ""

curl -X GET -H "Content-Type: application/json" \
	http://udr.free5gc.org:8000/nudr-dr/v2/application-data/influenceData?dnns=internet
echo -e "\n\n"

exit 0
