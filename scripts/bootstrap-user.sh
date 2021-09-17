#! /usr/bin/env bash

#
# Bootstrap a user's password
#
# Usage:
#     ./scripts/bootstrap-govcloud.sh <username>
#

set -eu -o pipefail

if [ "$#" -ne 1 ]; then
    echo "Illegal number of parameters. Must supply IAM username"
    exit 1
fi

USERNAME=$1
PASSWORD=$(aws secretsmanager get-random-password --password-length 24 --require-each-included-type | jq -r ".RandomPassword")
ALIAS=$(aws iam list-account-aliases | jq -r .AccountAliases[0])

if aws iam get-user --user-name "${USERNAME}" > /dev/null 2>&1 ; then
  aws iam create-login-profile --user-name "${USERNAME}" --password "${PASSWORD}"
else
  echo "User ${USERNAME} has not yet been provisioned"
  exit 1
fi

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
echo "1. Log in to the console with your temporary password."
echo "2. Create an MFA. Save MFA to 1Password as One Time Password (OTP)"
if [ "${AWS_REGION}" == "us-gov-east-1" ] ||  [ "${AWS_REGION}" == "us-gov-west-1" ]; then
  echo -e "\tURL: https://console.amazonaws-us-gov.com/iam/home?region=${AWS_REGION}#/users/${USERNAME}?section=security_credentials"
else
  echo -e "\tURL: https://console.aws.amazon.com/iam/home?#/users/${USERNAME}?section=security_credentials"
fi
echo "3. Log out of the AWS Console"
echo "4. Log in to the console with your new MFA"
echo "5. Reset your password (min 20 chars, requires upper and lowercase, numbers and symbols)"
echo "6. Assume the IAM Role you wish to use."
echo
echo "NOTE: You will not be able to do anything in the account unless you log in with MFA and assume the AWS IAM role for your project."
echo
