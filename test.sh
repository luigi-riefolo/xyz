#!/usr/bin/env bash

USER="luigi@luigi.com"
USER2="luigi2@luigi.com"
HOST="127.0.0.1:8080"

# create user
echo "Create user $USER"
curl -s -w "\n" -k -d '{
    "email": "'"$USER"'",
    "password": "secret",
    "firstname": "luigi",
    "lastname": "luigi"
}' https://$HOST/api/createUser | jq

# create user
echo -e "\nCreate user $USER2"
curl -s -w "\n" -k -d '{
    "email": "'"$USER2"'",
    "password": "secret",
    "firstname": "luigi",
    "lastname": "luigi"
}' https://$HOST/api/createUser | jq

# sign in
echo -e "\nSign in $USER"
TOKEN=$(curl -s \
"https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=$FIREBASE_API_KEY" \
-H 'Content-Type: application/json' \
--data-binary '{
    "email": "'"$USER"'",
    "password":"secret",
    "returnSecureToken":true
}' | jq -r '.idToken')

# create project
echo -e "\nCreate project for $USER"
PROJECT_RESP="$( \
curl -s -w "\n" -k --show-error --fail -H "Authorization: Bearer $TOKEN" -d '{
    "contributors": ["'"$USER2"'"]
}' https://$HOST/api/createProject)"
echo $PROJECT_RESP | jq

PROJECT_ID="$(echo $PROJECT_RESP | jq -r '.id')"

# add project contributors
echo -e "\nAdd project contributor $USER"
curl -s -w "\n" -k -H "Authorization: Bearer $TOKEN" -d '{
    "contributors": ["luigi3@luigi.com"]
}' https://$HOST/api/projects/$PROJECT_ID/addContributors | jq

# add project devices
echo -e "\nAdd project device"
curl -s -w "\n" -k -H "Authorization: Bearer $TOKEN" -d '{
    "devices": ["12345-XXX"]
}' https://$HOST/api/projects/$PROJECT_ID/addDevices | jq


# get project's devices
echo -e "\nGet project's devices"
curl -s -w "\n" -k -H "Authorization: Bearer $TOKEN" \
    https://$HOST/api/projects/$PROJECT_ID/devices | jq

# get project's devices
echo -e "\nGet projects"
curl -s -w "\n" -k -H "Authorization: Bearer $TOKEN" \
    https://$HOST/api/projects | jq
