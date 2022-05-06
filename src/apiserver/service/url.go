/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.,
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the ",License",); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an ",AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import (
	"bytes"
	"configcenter/src/apiserver/service/match"
	"configcenter/src/common/json"
	"errors"
	"fmt"
	"github.com/emicklei/go-restful"
	"io/ioutil"
	"regexp"
	"strings"
)

// URLPath url path filter
type URLPath string

// FilterChain url path filter
func (u URLPath) FilterChain(req *restful.Request) (RequestType, error) {
	var serverType RequestType
	var err error
	err = u.UrlTransfer(req)
	if err != nil {
		return UnknownType, err
	}
	switch {
	case u.WithCache(req):
		serverType = CacheType
	case u.WithTopo(req):
		serverType = TopoType
	case u.WithHost(req):
		serverType = HostType
	case u.WithProc(req):
		serverType = ProcType
	case u.WithEvent(req):
		serverType = EventType
	case u.WithDataCollect(req):
		return DataCollectType, nil
	case u.WithOperation(req):
		return OperationType, nil
	case u.WithTask(req):
		return TaskType, nil
	case u.WithAdmin(req):
		return AdminType, nil
	case u.WithCloud(req):
		return CloudType, nil
	default:
		if server, isHit := match.FilterMatch(req); isHit {
			return RequestType(server), nil
		}
		serverType = UnknownType
		err = errors.New("unknown requested with backend process")
	}

	return serverType, err
}

var topoURLRegexp = regexp.MustCompile(fmt.Sprintf("^/api/v3/(%s)/(inst|object|objects|topo|biz|module|set|resource)/.*$", verbs))
var objectURLRegexp = regexp.MustCompile(fmt.Sprintf("^/api/v3/(%s)/object$", verbs))

// WithTopo parse topo api's url
func (u *URLPath) WithTopo(req *restful.Request) (isHit bool) {
	topoRoot := "/topo/v3"
	from, to := rootPath, topoRoot
	switch {
	case strings.HasPrefix(string(*u), rootPath+"/biz/"):
		from, to, isHit = rootPath+"/biz", topoRoot+"/app", true

	case strings.HasPrefix(string(*u), rootPath+"/topo/"):
		from, to, isHit = rootPath, topoRoot, true

	case topoURLRegexp.MatchString(string(*u)):
		from, to, isHit = rootPath, topoRoot, true

	case objectURLRegexp.MatchString(string(*u)):
		from, to, isHit = rootPath, topoRoot, true

	case strings.HasPrefix(string(*u), rootPath+"/identifier/"):
		from, to, isHit = rootPath, topoRoot, true

	case strings.HasPrefix(string(*u), rootPath+"/inst/"):
		from, to, isHit = rootPath, topoRoot, true

	case strings.HasPrefix(string(*u), rootPath+"/module/"):
		from, to, isHit = rootPath, topoRoot, true

	case strings.HasPrefix(string(*u), rootPath+"/object/"):
		from, to, isHit = rootPath, topoRoot, true

	case strings.HasPrefix(string(*u), rootPath+"/set/"):
		from, to, isHit = rootPath, topoRoot, true

	case strings.Contains(string(*u), "/objectclassification"):
		from, to, isHit = rootPath, topoRoot, true

	case strings.Contains(string(*u), "/classificationobject"):
		from, to, isHit = rootPath, topoRoot, true

	case strings.Contains(string(*u), "/objectattr"):
		from, to, isHit = rootPath, topoRoot, true
	case strings.Contains(string(*u), "/objectunique"):
		from, to, isHit = rootPath, topoRoot, true

	case strings.Contains(string(*u), "/objectattgroup"):
		from, to, isHit = rootPath, topoRoot, true

	case strings.Contains(string(*u), "/objectattgroupproperty"):
		from, to, isHit = rootPath, topoRoot, true

	// TODO remove it
	case strings.Contains(string(*u), "/objectattgroupasst"):
		from, to, isHit = rootPath, topoRoot, true

	case strings.Contains(string(*u), "/objecttopo"):
		from, to, isHit = rootPath, topoRoot, true

	case strings.Contains(string(*u), "/topomodelmainline"):
		from, to, isHit = rootPath, topoRoot, true

	case strings.Contains(string(*u), "/topoinst"):
		from, to, isHit = rootPath, topoRoot, true

	case strings.Contains(string(*u), "/topopath"):
		from, to, isHit = rootPath, topoRoot, true

	case strings.Contains(string(*u), "/instassttopo"):
		from, to, isHit = rootPath, topoRoot, true

	case strings.Contains(string(*u), "/objecttopology"):
		from, to, isHit = rootPath, topoRoot, true

	case strings.Contains(string(*u), "/topoassociationtype"):
		from, to, isHit = rootPath, topoRoot, true

	case strings.Contains(string(*u), "/objectassociation"):
		from, to, isHit = rootPath, topoRoot, true

	case strings.Contains(string(*u), "/instassociation"):
		from, to, isHit = rootPath, topoRoot, true

	case strings.Contains(string(*u), "/insttopo"):
		from, to, isHit = rootPath, topoRoot, true

	case strings.Contains(string(*u), "/instance"):
		from, to, isHit = rootPath, topoRoot, true

	case strings.Contains(string(*u), "/instassociationdetail"):
		from, to, isHit = rootPath, topoRoot, true

	case strings.Contains(string(*u), "/associationtype"):
		from, to, isHit = rootPath, topoRoot, true

	case strings.Contains(string(*u), "/find/full_text"):
		from, to, isHit = rootPath, topoRoot, true

	case strings.Contains(string(*u), "/find/audit_dict"):
		from, to, isHit = rootPath, topoRoot, true

	case strings.Contains(string(*u), "/findmany/audit_list"):
		from, to, isHit = rootPath, topoRoot, true

	case strings.HasPrefix(string(*u), rootPath+"/find/audit"):
		from, to, isHit = rootPath, topoRoot, true

	case topoURLRegexp.MatchString(string(*u)):
		from, to, isHit = rootPath, topoRoot, true

	default:
		isHit = false
	}

	if isHit {
		u.revise(req, from, to)
		return true
	}
	return false
}

// hostCloudAreaURLRegexp host server operator cloud area api regex
var hostCloudAreaURLRegexp = regexp.MustCompile(fmt.Sprintf("^/api/v3/(%s)/(cloudarea|cloudarea/.*)$", verbs))
var hostURLRegexp = regexp.MustCompile(fmt.Sprintf("^/api/v3/(%s)/(host|hosts|host_apply_rule|host_apply_plan)/.*$", verbs))

// WithHost transform the host's url
func (u *URLPath) WithHost(req *restful.Request) (isHit bool) {
	hostRoot := "/host/v3"
	from, to := rootPath, hostRoot

	switch {
	case strings.HasPrefix(string(*u), rootPath+"/host/"):
		from, to, isHit = rootPath, hostRoot, true

	case strings.HasPrefix(string(*u), rootPath+"/hosts/"):
		from, to, isHit = rootPath, hostRoot, true

	// dynamic grouping URL matching, and proxy to host server.
	case string(*u) == (rootPath + "/dynamicgroup"):
		from, to, isHit = rootPath, hostRoot, true

	case strings.HasPrefix(string(*u), rootPath+"/dynamicgroup/"):
		from, to, isHit = rootPath, hostRoot, true

	case string(*u) == (rootPath + "/usercustom"):
		from, to, isHit = rootPath, hostRoot, true

	case strings.HasPrefix(string(*u), rootPath+"/usercustom/"):
		from, to, isHit = rootPath, hostRoot, true

	case hostCloudAreaURLRegexp.MatchString(string(*u)):
		from, to, isHit = rootPath, hostRoot, true

	case hostURLRegexp.MatchString(string(*u)):
		from, to, isHit = rootPath, hostRoot, true

	case strings.HasPrefix(string(*u), rootPath+"/system/config"):
		from, to, isHit = rootPath, hostRoot, true

	case strings.HasPrefix(string(*u), rootPath+"/findmany/module_relation/bk_biz_id/"):
		from, to, isHit = rootPath, hostRoot, true
	default:
		isHit = false
	}

	if isHit {
		u.revise(req, from, to)
		return true
	}
	return false
}

// WithEvent transform event's url
func (u *URLPath) WithEvent(req *restful.Request) (isHit bool) {
	eventRoot := "/event/v3"
	from, to := rootPath, eventRoot

	switch {
	case strings.HasPrefix(string(*u), rootPath+"/event/"):
		from, to, isHit = rootPath+"/event", eventRoot, true

	default:
		isHit = false
	}

	if isHit {
		u.revise(req, from, to)
		return true
	}
	return false
}

const verbs = "create|createmany|update|updatemany|delete|deletemany|find|findmany"

var procUrlRegexp = regexp.MustCompile(fmt.Sprintf("^/api/v3/(%s)/proc/.*$", verbs))

// WithProc transform the proc's url
func (u *URLPath) WithProc(req *restful.Request) (isHit bool) {
	procRoot := "/process/v3"
	from, to := rootPath, procRoot

	switch {
	case strings.HasPrefix(string(*u), rootPath+"/proc/"):
		from, to, isHit = rootPath+"/proc", procRoot, true
	case procUrlRegexp.MatchString(string(*u)):
		from, to, isHit = rootPath, procRoot, true
	default:
		isHit = false
	}

	if isHit {
		u.revise(req, from, to)
		return true
	}
	return false
}

// WithDataCollect transform DataCollect's url
func (u *URLPath) WithDataCollect(req *restful.Request) (isHit bool) {
	dataCollectRoot := "/collector/v3"
	from, to := rootPath, dataCollectRoot

	switch {
	case strings.HasPrefix(string(*u), rootPath+"/collector/"):
		from, to, isHit = rootPath+"/collector", dataCollectRoot, true

	default:
		isHit = false
	}

	if isHit {
		u.revise(req, from, to)
		return true
	}
	return false
}

var operationUrlRegexp = regexp.MustCompile(fmt.Sprintf("^/api/v3/(%s)/operation/.*$", verbs))

// WithOperation transform OperationStatistic's url
func (u *URLPath) WithOperation(req *restful.Request) (isHit bool) {
	statisticsRoot := "/operation/v3"
	from, to := rootPath, statisticsRoot

	switch {
	case strings.HasPrefix(string(*u), rootPath+"/operation/"):
		from, to, isHit = rootPath, statisticsRoot, true
	case operationUrlRegexp.MatchString(string(*u)):
		from, to, isHit = rootPath, statisticsRoot, true
	default:
		isHit = false
	}

	if isHit {
		u.revise(req, from, to)
		return true
	}
	return false
}

// WithTask transform task server  url
func (u *URLPath) WithTask(req *restful.Request) (isHit bool) {
	statisticsRoot := "/task/v3"
	from, to := rootPath, statisticsRoot

	switch {
	case strings.HasPrefix(string(*u), rootPath+"/task/"):
		from, to, isHit = rootPath, statisticsRoot, true

	default:
		isHit = false
	}

	if isHit {
		u.revise(req, from, to)
		return true
	}
	return false
}

// WithAdmin transform admin server url
func (u *URLPath) WithAdmin(req *restful.Request) (isHit bool) {
	adminRoot := "/migrate/v3"
	from, to := rootPath, adminRoot

	switch {
	case strings.HasPrefix(string(*u), rootPath+"/admin/"):
		from, to, isHit = rootPath+"/admin", adminRoot, true

	default:
		isHit = false
	}

	if isHit {
		u.revise(req, from, to)
		return true
	}
	return false
}

var cloudUrlRegexp = regexp.MustCompile(fmt.Sprintf("^/api/v3/(%s)/cloud/.*$", verbs))

// WithCloud transform cloud's url
func (u *URLPath) WithCloud(req *restful.Request) (isHit bool) {
	cloudRoot := "/cloud/v3"
	from, to := rootPath, cloudRoot

	switch {
	case strings.HasPrefix(string(*u), rootPath+"/cloud/"):
		from, to, isHit = rootPath, cloudRoot, true
	case cloudUrlRegexp.MatchString(string(*u)):
		from, to, isHit = rootPath, cloudRoot, true
	default:
		isHit = false
	}

	if isHit {
		u.revise(req, from, to)
		return true
	}
	return false
}

func (u URLPath) revise(req *restful.Request, from, to string) {
	req.Request.RequestURI = to + req.Request.RequestURI[len(from):]
	req.Request.URL.Path = to + req.Request.URL.Path[len(from):]
}

// WithCache transform cache service's url
func (u *URLPath) WithCache(req *restful.Request) (isHit bool) {
	cacheRoot := "/cache/v3"
	from, to := rootPath, cacheRoot

	switch {
	case strings.HasPrefix(string(*u), rootPath+"/cache/"):
		from, to, isHit = rootPath+"/cache", cacheRoot, true
	default:
		isHit = false
	}

	if isHit {
		u.revise(req, from, to)
		return true
	}
	return false
}

type ObjIdParam struct {
	BkObjId string `json:"bk_obj_id"`
}

type ObjInstIdParam struct {
	BkObjId  string `json:"bk_obj_id"`
	BkInstId int    `json:"bk_inst_id"`
}

type AsstIdParam struct {
	Id int `json:"id"`
}

func (u *URLPath) UrlTransfer(req *restful.Request) error {

	body, err := ioutil.ReadAll(req.Request.Body)
	req.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	req.Request.Body.Close()
	if err != nil {
		err = errors.New("read request body failed")
		return err
	}
	switch *u {
	case "/api/v3/cc/search_inst/":
		/*
		  name: search_inst
		  label: 根据关联关系实例查询模型实例
		  label_en: search insts by condition
		  suggest_method: POST
		  api_type: query
		*/
		ObjId := new(ObjIdParam)
		err = json.Unmarshal(body, &ObjId)
		if err != nil {
			err = errors.New("request body cannot be empty")
			return err
		}
		*u = "/api/v3/find/instassociation/object"
		req.Request.RequestURI = "/api/v3/find/instassociation/object/" + ObjId.BkObjId
		req.Request.URL.Path = "/api/v3/find/instassociation/object/" + ObjId.BkObjId

	case "/api/v3/cc/create_inst/":
		/*
		  name: create_inst
		  label: 创建实例
		  label_en: create a new inst
		  suggest_method: POST
		  api_type: operate
		*/
		ObjId := new(ObjIdParam)
		err = json.Unmarshal(body, &ObjId)
		if err != nil {
			err = errors.New("request body cannot be empty")
			return err
		}
		*u = "/api/v3/create/instance/object"
		req.Request.RequestURI = "/api/v3/create/instance/object/" + ObjId.BkObjId
		req.Request.URL.Path = "/api/v3/create/instance/object/" + ObjId.BkObjId

	case "/api/v3/cc/update_inst/":
		/*
		  name: update_inst
		  label: 更新对象实例
		  label_en: update a inst
		  suggest_method: POST
		  api_type: operate
		*/
		ObjInstId := new(ObjInstIdParam)
		err = json.Unmarshal(body, &ObjInstId)
		if err != nil {
			err = errors.New("request body cannot be empty")
			return err
		}
		*u = "/api/v3/update/instance/object"
		req.Request.RequestURI = fmt.Sprintf("/api/v3/update/instance/object/%s/inst/%d", ObjInstId.BkObjId, ObjInstId.BkInstId)
		req.Request.URL.Path = fmt.Sprintf("/api/v3/update/instance/object/%s/inst/%d", ObjInstId.BkObjId, ObjInstId.BkInstId)

	case "/api/v3/cc/batch_update_inst/":
		/*
		  name: batch_update_inst
		  label: 批量更新对象实例
		  label_en: update inst batch
		  suggest_method: POST
		  api_type: operate
		*/
		ObjId := new(ObjIdParam)
		err = json.Unmarshal(body, &ObjId)
		if err != nil {
			err = errors.New("request body cannot be empty")
			return err
		}
		*u = "/api/v3/updatemany/instance/object"
		req.Request.RequestURI = "/api/v3/updatemany/instance/object/" + ObjId.BkObjId
		req.Request.URL.Path = "/api/v3/updatemany/instance/object/" + ObjId.BkObjId

	case "/api/v3/cc/delete_inst/":
		/*
		  name: delete_inst
		  label: 删除实例
		  label_en: delete a inst
		  suggest_method: POST
		  api_type: operate
		*/
		ObjInstId := new(ObjInstIdParam)
		err = json.Unmarshal(body, &ObjInstId)
		if err != nil {
			err = errors.New("request body cannot be empty")
			return err
		}
		*u = "/api/v3/delete/instance/object"
		req.Request.RequestURI = fmt.Sprintf("/api/v3/delete/instance/object/%s/inst/%d", ObjInstId.BkObjId, ObjInstId.BkInstId)
		req.Request.URL.Path = fmt.Sprintf("/api/v3/delete/instance/object/%s/inst/%d", ObjInstId.BkObjId, ObjInstId.BkInstId)

	case "/api/v3/cc/list_hosts_without_biz/":
		/*
		  name: list_hosts_without_biz
		  label: 没有业务ID的主机查询
		  label_en: list host without business id
		  suggest_method: POST
		  api_type: operate
		*/
		*u = "/api/v3/hosts/list_hosts_without_app"
		req.Request.RequestURI = "/api/v3/hosts/list_hosts_without_app"
		req.Request.URL.Path = "/api/v3/hosts/list_hosts_without_app"

	case "/api/v3/cc/batch_update_host/":
		/*
		  name: batch_update_host
		  label: 批量更新主机属性
		  label_en: update host batch
		  suggest_method: POST
		  api_type: operate
		*/
		*u = "/api/v3/hosts/property/batch"
		req.Request.RequestURI = "/api/v3/hosts/property/batch"
		req.Request.URL.Path = "/api/v3/hosts/property/batch"

	case "/api/v3/cc/search_hostidentifier/":
		/*
		  name: search_hostidentifier
		  label: 根据条件查询主机身份
		  label_en: search host identifier
		  suggest_method: POST
		  api_type: query
		*/
		*u = "/api/v3/identifier/host/search"
		req.Request.RequestURI = "/api/v3/identifier/host/search"
		req.Request.URL.Path = "/api/v3/identifier/host/search"

	case "/api/v3/cc/find_instance_association/":
		/*
		  name: find_instance_association
		  label: 查询模型实例之间的关联关系
		  label_en: find association between object's instance
		  suggest_method: POST
		  api_type: query
		*/
		*u = "/api/v3/find/instassociation"
		req.Request.RequestURI = "/api/v3/find/instassociation"
		req.Request.URL.Path = "/api/v3/find/instassociation"

	case "/api/v3/cc/add_instance_association/":
		/*
		  name: add_instance_association
		  label: 新建模型实例之间的关联关系
		  label_en: add association between object's instance
		  suggest_method: POST
		  api_type: query
		*/
		*u = "/api/v3/create/instassociation"
		req.Request.RequestURI = "/api/v3/create/instassociation"
		req.Request.URL.Path = "/api/v3/create/instassociation"

	case "/api/v3/cc/delete_instance_association/":
		/*
		  name: delete_instance_association
		  label: 删除模型实例之间的关联关系
		  label_en: delete association between object's instance
		  suggest_method: DELETE
		  api_type: query
		*/
		Id := new(AsstIdParam)
		err = json.Unmarshal(body, &Id)
		if err != nil {
			err = errors.New("request body cannot be empty")
			return err
		}
		*u = "/api/v3/delete/instassociation"
		req.Request.RequestURI = fmt.Sprintf("/api/v3/delete/instassociation/%d", Id.Id)
		req.Request.URL.Path = fmt.Sprintf("/api/v3/delete/instassociation/%d", Id.Id)

	case "/api/v3/cc/search_related_inst_asso/":
		/*
		  name: search_related_inst_asso
		  label: 查询某实例所有的关联关系（包含其作为关联关系原模型和关联关系目标模型的情况）
		  label_en: search a instance's all associations, including associations which is self associated or being associated
		  suggest_method: POST
		  api_type: query
		*/
		*u = "/api/v3/find/instassociation/related"
		req.Request.RequestURI = "/api/v3/find/instassociation/related"
		req.Request.URL.Path = "/api/v3/find/instassociation/related"

	case "/api/v3/cc/delete_related_inst_asso/":
		/*
		  name: delete_related_inst_asso
		  label: 删除某实例所有的关联关系（包含其作为关联关系原模型和关联关系目标模型的情况）
		  label_en: delete all associations of an instance (including cases that the instance is the association source  and it is the Association target)
		  suggest_method: POST
		  api_type: query
		*/
		*u = "/api/v3/delete/instassociation/batch"
		req.Request.RequestURI = "/api/v3/delete/instassociation/batch"
		req.Request.URL.Path = "/api/v3/delete/instassociation/batch"

	default:
		return nil
	}

	return nil
}
