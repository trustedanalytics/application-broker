#!/usr/bin/env bash
###############################################################################
#                           INTEL CONFIDENTIAL
#                 Copyright 2015 - 2016 Intel Corporation
#
# The source code contained or described herein and all documents
# related to the source code ("Material") are owned by Intel Corporation
# or its suppliers or licensors. Title to the Material remains with
# Intel Corporation or its suppliers and licensors. The Material contains
# trade secrets and proprietary and confidential information of Intel
# or its suppliers and licensors. The Material is protected by worldwide
# copyright and trade secret laws and treaty provisions. No part of
# the Material may be used, copied, reproduced, modified, published,
# uploaded, posted, transmitted, distributed, or disclosed in any way
# without Intel's prior express written permission.
#
# No license under any patent, copyright, trade secret or other intellectual
# property right is granted to or conferred upon you by disclosure
# or delivery of the Materials, either expressly, by implication, inducement,
# estoppel or otherwise. Any license under such intellectual property rights
# must be express and approved by Intel in writing.
###############################################################################

function run_tests_in {
    cd ./${1}
    go test -race -coverprofile=coverage.out
    if [[ $? -eq 0 ]]
    then
        go tool cover -func=coverage.out
        go tool cover -html=coverage.out    # Require xdg-open from xdg-utils to open default browser
        rm coverage.out
    fi
    cd ..
}

run_tests_in broker
run_tests_in cloud
