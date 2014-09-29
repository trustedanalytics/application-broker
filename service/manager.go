package service

import (
/*
	"errors"
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/intel-data/types-cf"
	"net/http"
*/
)

type ServiceManager struct {
	//client *rabbithole.Client
}

/*
func newRabbitAdmin(brokerUrl, username, password string) (*rabbitAdmin, error) {
	client, err := rabbithole.NewClient(brokerUrl, username, password)
	if err != nil {
		return nil, err
	}
	return &rabbitAdmin{client}, nil
}

func (a *rabbitAdmin) isVhost(username string) (bool, error) {
	info, err := a.client.GetVhost(username)
	if info != nil {
		return true, nil
	} else if err.Error() == "not found" { // TODO: Create PR to expose the 404 in more user-friendly way
		return false, nil
	}
	return false, &rabbitAdminError{broker.ErrCodeOther, err}
}

func (a *rabbitAdmin) createVhost(vhostname string, tracing bool) error {
	if found, err := a.isVhost(vhostname); err != nil {
		return err
	} else if found {
		msg := fmt.Sprintf("Virtual host already exists: [%v]", vhostname)
		return &rabbitAdminError{broker.ErrCodeConflict, errors.New(msg)}
	}

	settings := rabbithole.VhostSettings{tracing}
	resp, err := a.client.PutVhost(vhostname, settings)
	if err != nil {
		return &rabbitAdminError{broker.ErrCodeOther, err}
	}
	return checkResponseAndClose(resp)
}

func (a *rabbitAdmin) deleteVhost(vhostname string) error {
	resp, err := a.client.DeleteVhost(vhostname)
	if err != nil {
		return &rabbitAdminError{broker.ErrCodeOther, err}
	}
	return checkResponseAndClose(resp)
}

func (a *rabbitAdmin) isUser(username string) (bool, error) {
	info, err := a.client.GetUser(username)
	if info != nil {
		return true, nil
	} else if err.Error() == "not found" {
		return false, nil
	}
	return false, &rabbitAdminError{broker.ErrCodeOther, err}
}

func (a *rabbitAdmin) createUser(username, password string) error {
	if found, err := a.isUser(username); err != nil {
		return err
	} else if found {
		msg := fmt.Sprintf("User already exists: %v", username)
		return &rabbitAdminError{broker.ErrCodeConflict, errors.New(msg)}
	}

	settings := rabbithole.UserSettings{
		Name:     username,
		Password: password,
		Tags:     "management",
	}
	resp, err := a.client.PutUser(username, settings)
	if err != nil {
		return &rabbitAdminError{broker.ErrCodeOther, err}
	}
	return checkResponseAndClose(resp)
}

func (a *rabbitAdmin) deleteUser(username string) error {
	resp, err := a.client.DeleteUser(username)
	if err != nil {
		return &rabbitAdminError{broker.ErrCodeOther, err}
	}
	return checkResponseAndClose(resp)
}

func (a *rabbitAdmin) grantAllPermissionsIn(username, vhostname string) error {
	unlimited := rabbithole.Permissions{".*", ".*", ".*"}
	resp, err := a.client.UpdatePermissionsIn(vhostname, username, unlimited)
	if err != nil {
		return &rabbitAdminError{broker.ErrCodeOther, err}
	}
	return checkResponseAndClose(resp)
}

func checkResponseAndClose(resp *http.Response) error {
	defer resp.Body.Close()

	switch code := resp.StatusCode; code {
	case http.StatusOK,
		http.StatusCreated,
		http.StatusAccepted,
		http.StatusNoContent:
		return nil
	case http.StatusNotFound:
		err := errors.New("Entity not found")
		return &rabbitAdminError{broker.ErrCodeGone, err}
	default:
		err := errors.New(fmt.Sprintf("Unexpected response received: [%v]", code))
		return &rabbitAdminError{broker.ErrCodeOther, err}
	}
}
*/
