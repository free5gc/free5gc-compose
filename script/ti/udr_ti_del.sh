set -e

curl -X DELETE -H "Content-Type: application/json" \
	http://udr.free5gc.org:8000/nudr-dr/v1/application-data/influenceData/914d1e66-15ae-4b4a-8b5e-d397b510ea93

exit 0
