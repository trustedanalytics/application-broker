/**
 * Copyright (c) 2015 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cloud

import (
	log "github.com/cihub/seelog"
	"github.com/trustedanalytics/go-cf-lib/types"
)

type Transaction struct {
	components []types.Component
}

func NewTransaction() *Transaction {
	tr := Transaction{
		components: make([]types.Component, 0),
	}
	return &tr
}

func (t *Transaction) AddApplication(app *types.CfAppResource) {
	comp := types.Component{
		GUID: app.Meta.GUID,
		Name: app.Entity.Name,
		Type: types.ComponentApp,
	}
	t.components = append(t.components, comp)
}

func (t *Transaction) AddComponentClone(clone *types.ComponentClone) {
	if clone != nil && len(clone.CloneGUID) > 0 {
		comp := types.Component{
			GUID: clone.CloneGUID,
			Name: clone.Component.Name,
			Type: clone.Component.Type,
		}
		t.components = append(t.components, comp)
	}
}

func (t *Transaction) Rollback(cloud *CloudAPI) {
	log.Errorf("Aborting transaction. Deprovisioning already spawned components")
	cloud.deprovisionComponents(t.components)
}
