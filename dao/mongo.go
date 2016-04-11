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
	"github.com/trustedanalytics/application-broker/service/extension"
	"github.com/trustedanalytics/go-cf-lib/types"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Mongo struct {
	session *mgo.Session
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
	return &Mongo{session: session}
}

func (c *Mongo) Get() ([]*extension.ServiceExtension, error) {
	result := []*extension.ServiceExtension{}
	session := c.session.Copy()
	defer session.Close()
	services := session.DB("").C("services")
	err := services.Find(nil).All(&result)
	if err != nil {
		log.Errorf("Problems while getting catalog: [%v]", err)
		return nil, types.InternalServerError{Context: "Could not get catalog from DB"}
	}
	return result, nil
}

func (c *Mongo) Find(id string) (*extension.ServiceExtension, error) {
	result := new(extension.ServiceExtension)
	session := c.session.Copy()
	defer session.Close()
	services := session.DB("").C("services")
	err := services.Find(bson.M{"service.id": id}).One(&result)
	if err != nil || result == nil {
		log.Errorf("No service found in catalog for id: [%v], Err: [%v]", id, err)
		return nil, types.ServiceNotFoundError{}
	}
	return result, nil
}

func (c *Mongo) Append(service *extension.ServiceExtension) error {
	session := c.session.Copy()
	defer session.Close()
	services := session.DB("").C("services")
	count, _ := services.Find(bson.M{"service.id": service.ID}).Count()
	if count > 0 {
		log.Errorf("Service already exists in catalog for id: [%v]", service.ID)
		return types.ServiceAlreadyExistsError{}
	}
	count, _ = services.Find(bson.M{"service.name": service.Name}).Count()
	if count > 0 {
		log.Errorf("Service already exists in catalog for name: [%v]", service.Name)
		return types.ServiceAlreadyExistsError{}
	}

	err := services.Insert(service)
	if err != nil {
		log.Errorf("Could not insert service to catalog: [%v]", err)
		err = types.InternalServerError{Context: "Problem while appending service to DB"}
	}
	return err
}

func (c *Mongo) Update(service *extension.ServiceExtension) error {
	session := c.session.Copy()
	defer session.Close()
	services := session.DB("").C("services")

	err := services.Update(bson.M{"service.id": service.ID}, service)
	if err != nil {
		log.Errorf("Could not insert service to catalog: [%v]", err)
		err = types.InternalServerError{Context: "Problem while appending service to DB"}
	}
	return err
}

func (c *Mongo) Remove(serviceID string) error {
	session := c.session.Copy()
	defer session.Close()
	services := session.DB("").C("services")

	if err := services.Remove(bson.M{"service.id": serviceID}); err != nil {
		log.Errorf("Could not remove service from catalog: [%v]", err)
		return types.InternalServerError{Context: "Problem while removing service from DB"}
	}
	return nil
}

func (c *Mongo) AppendInstance(instance extension.ServiceInstanceExtension) error {
	session := c.session.Copy()
	defer session.Close()
	instances := session.DB("").C("instances")

	err := instances.Insert(instance)
	if err != nil {
		log.Errorf("Could not insert instance to database: [%v]", err)
		return types.InternalServerError{Context: "Problem with storing svc instance in DB"}
	}
	return nil
}

func (c *Mongo) FindInstance(id string) (*extension.ServiceInstanceExtension, error) {
	session := c.session.Copy()
	defer session.Close()
	instances := session.DB("").C("instances")

	result := new(extension.ServiceInstanceExtension)
	instances.Find(bson.M{"id": id}).One(&result)
	if len(result.ID) == 0 {
		log.Errorf("No service instance found in database for id: [%v]", id)
		return nil, types.InstanceNotFoundError{}
	}
	return result, nil
}

func (c *Mongo) HasInstancesOf(serviceID string) (bool, error) {
	session := c.session.Copy()
	defer session.Close()
	instances := session.DB("").C("instances")

	count, err := instances.Find(bson.M{"serviceid": serviceID}).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (c *Mongo) RemoveInstance(id string) error {
	session := c.session.Copy()
	defer session.Close()
	instances := session.DB("").C("instances")

	err := instances.Remove(bson.M{"id": id})
	if err != nil {
		log.Errorf("Could not delete instance %v from database: [%v]", id, err)
		err = types.InternalServerError{Context: err.Error()}
	}
	return err
}
