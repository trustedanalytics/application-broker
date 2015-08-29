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

package dao

import (
	log "github.com/cihub/seelog"
	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/trustedanalytics/application-broker/misc"
	"github.com/trustedanalytics/application-broker/types"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Mongo struct {
	services  *mgo.Collection
	instances *mgo.Collection
}

func MongoFactory(envs *cfenv.App) *Mongo {
	uri := "localhost/application-broker"
	if envs != nil {
		instances, err := envs.Services.WithTag("mongodb")
		uri = instances[0].Credentials["uri"].(string)

		if err != nil {
			log.Critical("Running in cloud but no mongodb found")
		}
	}

	log.Infof("Connecting with mongodb: [%v]", uri)
	session, err := mgo.Dial(uri)
	if err != nil {
		log.Criticalf("Cannot dial to mongodb! Err: [%v]", err)
	}
	svcs := session.DB("").C("services")
	ists := session.DB("").C("instances")
	return &Mongo{services: svcs, instances: ists}
}

func (c *Mongo) Get() ([]*types.ServiceExtension, error) {
	result := []*types.ServiceExtension{}
	err := c.services.Find(nil).All(&result)
	if err != nil {
		log.Errorf("Problems while getting catalog: [%v]", err)
		return nil, misc.InternalServerError{Context: "Could not get catalog from DB"}
	}
	return result, nil
}

func (c *Mongo) Find(id string) (*types.ServiceExtension, error) {
	result := new(types.ServiceExtension)
	err := c.services.Find(bson.M{"service.id": id}).One(&result)
	if err != nil || result == nil {
		log.Errorf("No service found in catalog for id: [%v], Err: [%v]", id, err)
		return nil, misc.ServiceNotFoundError{}
	}
	return result, nil
}

func (c *Mongo) Append(service *types.ServiceExtension) error {
	count, _ := c.services.Find(bson.M{"service.id": service.ID}).Count()
	if count > 0 {
		log.Errorf("Service already exists in catalog for id: [%v]", service.ID)
		return misc.ServiceAlreadyExistsError{}
	}
	count, _ = c.services.Find(bson.M{"service.name": service.Name}).Count()
	if count > 0 {
		log.Errorf("Service already exists in catalog for name: [%v]", service.Name)
		return misc.ServiceAlreadyExistsError{}
	}

	err := c.services.Insert(service)
	if err != nil {
		log.Errorf("Could not insert service to catalog: [%v]", err)
		err = misc.InternalServerError{Context: "Problem while appending service to DB"}
	}
	return err
}

func (c *Mongo) Remove(serviceID string) error {
	if err := c.services.Remove(bson.M{"service.id": serviceID}); err != nil {
		log.Errorf("Could not remove service from catalog: [%v]", err)
		return misc.InternalServerError{Context: "Problem while removing service from DB"}
	}
	return nil
}

func (c *Mongo) AppendInstance(instance types.ServiceInstanceExtension) error {
	err := c.instances.Insert(instance)
	if err != nil {
		log.Errorf("Could not insert instance to database: [%v]", err)
		return misc.InternalServerError{Context: "Problem with storing svc instance in DB"}
	}
	return nil
}

func (c *Mongo) FindInstance(id string) (*types.ServiceInstanceExtension, error) {
	result := new(types.ServiceInstanceExtension)
	c.instances.Find(bson.M{"id": id}).One(&result)
	if len(result.ID) == 0 {
		log.Errorf("No service instance found in database for id: [%v]", id)
		return nil, misc.InstanceNotFoundError{}
	}
	return result, nil
}

func (c *Mongo) HasInstancesOf(serviceID string) (bool, error) {
	count, err := c.instances.Find(bson.M{"serviceid": serviceID}).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (c *Mongo) RemoveInstance(id string) error {
	err := c.instances.Remove(bson.M{"id": id})
	if err != nil {
		log.Errorf("Could not delete instance %v from database: [%v]", id, err)
		err = misc.InternalServerError{Context: err.Error()}
	}
	return err
}
