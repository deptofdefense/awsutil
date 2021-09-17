#! /usr/bin/env bash

#
# Reset a user's password
#
# Usage:
#     ./scripts/reset-user-password.sh <username>
#

set -eu -o pipefail

if [ "$#" -ne 1 ]; then
    echo "Illegal number of parameters. Must supply IAM username"
    exit 1
fi

USERNAME=$1
PASSWORD=$(aws secretsmanager get-random-password --password-length 24 --require-each-included-type | jq -r ".RandomPassword")
ALIAS=$(aws iam list-account-aliases | jq -r .AccountAliases[0])

aws iam update-login-profile --user-name "${USERNAME}" --password "${PASSWORD}"

echo
if [ "${AWS_REGION}" == "us-gov-east-1" ] ||  [ "${AWS_REGION}" == "us-gov-west-1" ]; then
  echo "Login URL: https://${ALIAS}.signin.amazonaws-us-gov.com/console"
else
  echo "Login URL: https://${ALIAS}.signin.aws.amazon.com/console"
fi

echo "Username: ${USERNAME}"
echo "Password: ${PASSWORD}"
echo
echo "Please follow these steps:"
echo "1. Log in to the console with your new password"
echo "2. Reset your password"
echo
