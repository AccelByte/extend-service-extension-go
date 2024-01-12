#!/usr/bin/env bash

set -e
set -o pipefail
#set -x

test -n "$AB_CLIENT_ID" || (echo "AB_CLIENT_ID is not set"; exit 1)
test -n "$AB_CLIENT_SECRET" || (echo "AB_CLIENT_SECRET is not set"; exit 1)
test -n "$AB_NAMESPACE" || (echo "AB_NAMESPACE is not set"; exit 1)

GUILD_ID='63d5802dd554c87f4e0b0707ae2f0af44c6f7d08f1ff3dec21a02728b10476e4'

function api_curl()
{
  curl -s -o http_response.out -w '%{http_code}' "$@" > http_code.out
  echo >> http_response.out
  cat http_response.out
}

echo 'Logging in client ...'

ACCESS_TOKEN="$(api_curl ${AB_BASE_URL}/iam/v3/oauth/token \
    -H 'Content-Type: application/x-www-form-urlencoded' \
    -u "$AB_CLIENT_ID:$AB_CLIENT_SECRET" \
    -d "grant_type=client_credentials" | jq --raw-output .access_token)"

if [ "$ACCESS_TOKEN" == "null" ]; then
    cat http_response.out
    exit 1
fi

echo 'Updating guild progression ...'

api_curl -X 'POST' \
  "http://localhost:8000/guild/v1/admin/namespace/$AB_NAMESPACE/progress" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d "{\"guildProgress\":{\"guildId\":\"$GUILD_ID\",\"namespace\":\"$AB_NAMESPACE\",\"objectives\":{\"additionalProp1\":0,\"additionalProp2\":0,\"additionalProp3\":0}}}"
echo
echo

if ! cat http_code.out | grep -q '\(200\|201\|204\|302\)'; then
    exit 1
fi

echo 'Getting guild progression ...'

curl -X 'GET' \
  "http://localhost:8000/guild/v1/admin/namespace/$AB_NAMESPACE/progress/$GUILD_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'accept: application/json'
echo
echo

if ! cat http_code.out | grep -q '\(200\|201\|204\|302\)'; then
    exit 1
fi
