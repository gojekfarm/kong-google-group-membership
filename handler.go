package main

import (
	"fmt"
	"sync"

	"github.com/Kong/go-pdk"
	"github.com/gojekfarm/kong-google-group-membership/groups"
	admin "google.golang.org/api/admin/directory/v1"
)

type Handler struct {
	mu              sync.Mutex // guards balance
	adminService    *admin.Service
	CredentialsPath string
	GroupAdmin      string
	GroupEmail      string
}

func New() interface{} {
	return &Handler{}
}

func (conf Handler) directoryService() (*admin.Service, error) {
	if conf.adminService == nil {
		conf.mu.Lock()
		adminSvc, err := groups.CreateDirectoryService(conf.GroupAdmin, conf.CredentialsPath)
		if err != nil {
			return nil, err
		}
		conf.adminService = adminSvc
		conf.mu.Unlock()
	}
	return conf.adminService, nil
}

func (conf Handler) Access(kong *pdk.PDK) {
	consumer, err := kong.Nginx.AskMap("ngx.ctx.authenticated_consumer")
	if err != nil {
		kong.Log.Err(err.Error())
		kong.Log.Err("This plugin depends on oidc plugin")
		return
	}

	adminService, err := conf.directoryService()
	if err != nil {
		kong.Log.Err(err.Error())
		return
	}
	isMember, err := adminService.Members.HasMember(conf.GroupEmail, fmt.Sprintf("%v", consumer["username"])).Do()

	if err != nil {
		kong.Log.Err(fmt.Sprintf("Error calling google admin directory service for consumer %v", consumer))
		kong.Log.Err(err.Error())
		return
	}

	if isMember.IsMember {
		kong.Log.Debug(fmt.Sprintf("membership: %v", conf.GroupEmail))
		kong.ServiceRequest.SetHeader("x-group-member", conf.GroupEmail)
	} else {
		kong.Nginx.Ask("kong.response.exit", 401)
	}
}
