#
# Copyright (c) 2016 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# Builds an artifact that can be used in offline deployment of the application.

set -e

VERSION=$(grep current_version .bumpversion.cfg | cut -d " " -f 3)
PROJECT_NAME=$(basename $(pwd))

# build project
cd Godeps/_workspace
mkdir -p src/github.com/trustedanalytics/
cd src/github.com/trustedanalytics/
ln -s ../../../../.. $PROJECT_NAME
cd ../../../../..

GOPATH=`godep path`:$GOPATH go test ./...
godep go build

rm Godeps/_workspace/src/github.com/trustedanalytics/$PROJECT_NAME
godep go clean

# assemble the artifact
PACKAGE_CATALOG=${PROJECT_NAME}-${VERSION}

# prepare build manifest
echo "commit_sha=$(git rev-parse HEAD)" > build_info.ini

# create zip package
zip -r ${PROJECT_NAME}-${VERSION}.zip * -x ${PROJECT_NAME}-${VERSION}.zip


echo "Zip package for $PROJECT_NAME project in version $VERSION has been prepared."
