#!/usr/bin/env bash
# Copyright 2018 Adobe
# All Rights Reserved.

# NOTICE: Adobe permits you to use, modify, and distribute this file in
# accordance with the terms of the Adobe license agreement accompanying
# it. If you have received this file from a source other than Adobe,
# then your use, modification, or distribution of it requires the prior
# written permission of Adobe. 

package_name='slh'
package='./src/github.com/adobe/sledgehammer/slh'

platforms=(${1//;/ })

for platform in "${platforms[@]}" 
do 
    platform_split=(${platform//-/ })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name='./bin/'$package_name'-'$GOOS'-'$GOARCH
    env CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build -ldflags "-X github.com/adobe/sledgehammer/slh/version.Version=$2 -X github.com/adobe/sledgehammer/slh/version.BuildDate=$3 -X github.com/adobe/sledgehammer/slh/version.GitCommit=$4" -o $output_name $package
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    else
        echo "Built $output_name"
    fi
done