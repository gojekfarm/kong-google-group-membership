package main

import (
	b64 "encoding/base64"
	"encoding/json"
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

type xUserInfo struct {
	Id string `json:"id"`
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
	encodedUserInfo, err := kong.Request.GetHeader("X-Userinfo")

	if err != nil {
		kong.Log.Err(err.Error())
		kong.Log.Err("This plugin depends on oidc plugin: missing header")
		return
	}
	userInfoString, err := b64.StdEncoding.DecodeString(encodedUserInfo)
	if err != nil {
		kong.Log.Err(err.Error())
		kong.Log.Err("This plugin depends on oidc plugin: X-Userinfo was not base64 encoded")
		return
	}
	userInfo := &xUserInfo{}
	if err := json.Unmarshal(userInfoString, userInfo); err != nil {
		kong.Log.Err(err.Error())
		kong.Log.Err("This plugin depends on oidc plugin: X-Userinfo was incorrectly set")
		return
	}

	adminService, err := conf.directoryService()
	if err != nil {
		kong.Log.Err(err.Error())
		return
	}
	isMember, err := adminService.Members.HasMember(conf.GroupEmail, userInfo.Id).Do()

	if err != nil {
		kong.Log.Err(fmt.Sprintf("Error calling google admin directory service for consumer %v", userInfo.Id))
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
