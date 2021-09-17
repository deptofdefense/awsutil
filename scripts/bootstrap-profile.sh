#!/bin/bash

mkdir -p ~/.aws

printf "[profile %s]\nregion=us-west-2\noutput=json\n" "${AWS_PROFILE}" >> ~/.aws/config
