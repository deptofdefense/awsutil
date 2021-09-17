#! /usr/bin/env bash

#
# Bootstrap the first user for a govcloud account
#
# These instructions are from http://govcloudconsolesetup.s3-us-gov-west-1.amazonaws.com/setup.html
#
# Usage:
#     ./scripts/bootstrap-govcloud.sh <profile> <username>
#

set -eu -o pipefail

if [ "$#" -ne 2 ]; then
    echo "Illegal number of parameters"
    exit 1
fi

PROFILE=$1
USERNAME=$2

GROUP_NAME="admin"
POLICY_NAME="admin"

echo "Create Group ${GROUP_NAME}"
aws --profile "${PROFILE}" iam create-group --group-name "${GROUP_NAME}" --path /
echo "Create Policy ${POLICY_NAME} in Group ${GROUP_NAME}"
aws --profile "${PROFILE}" iam put-group-policy --group-name "${GROUP_NAME}" --policy-name "${POLICY_NAME}" --policy-document "{ \"Version\": \"2012-10-17\", \"Statement\": [ { \"Effect\": \"Allow\", \"Action\": \"*\", \"Resource\": \"*\" } ] }"
echo "Create User ${USERNAME}"
aws --profile "${PROFILE}" iam create-user --user-name "${USERNAME}" --path /
echo "Add User ${USERNAME} to Group ${GROUP_NAME}"
aws --profile "${PROFILE}" iam add-user-to-group --group-name "${GROUP_NAME}" --user-name "${USERNAME}"
echo "List Users"
aws --profile "${PROFILE}" iam list-users --path-prefix /

echo -n "Enter a new password for user ${USERNAME}: "
read -rs PASSWORD

echo
echo "Set password for user ${USERNAME}"
aws --profile "${PROFILE}" iam create-login-profile --user-name "${USERNAME}" --password "${PASSWORD}"
echo "COMPLETE!"
