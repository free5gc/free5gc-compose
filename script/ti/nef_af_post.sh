set -e

curl -X POST -H "Content-Type: application/json" --data @./nef_ti_anyUE.json \
	http://nef.free5gc.org:8000/3gpp-traffic-influence/v1/af001/subscriptions

exit 0
