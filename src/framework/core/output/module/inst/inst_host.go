/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package inst

import (
	"strconv"
	"strings"

	cccommon "configcenter/src/common"
	"configcenter/src/framework/common"
	"configcenter/src/framework/core/errors"
	"configcenter/src/framework/core/log"
	"configcenter/src/framework/core/output/module/client"
	"configcenter/src/framework/core/output/module/model"
	"configcenter/src/framework/core/types"
)

var _ HostInterface = (*host)(nil)

// HostInterface the host interface
type HostInterface interface {
	IsExists() (bool, error)
	Create() error
	Update() error
	Save() error

	SetBusinessID(bizID int64)
	SetModuleIDS(moduleIDS []int64)

	GetModel() model.Model

	GetInstID() (int, error)
	GetInstName() string

	SetValue(key string, value interface{}) error
	GetValues() (types.MapStr, error)
}

type host struct {
	bizID     int64
	moduleIDS []int64
	target    model.Model
	datas     types.MapStr
}

func (cli *host) SetBusinessID(bizID int64) {
	cli.bizID = bizID
}

func (cli *host) SetModuleIDS(moduleIDS []int64) {
	cli.moduleIDS = moduleIDS
}

func (cli *host) GetModel() model.Model {
	return cli.target
}

func (cli *host) GetInstID() (int, error) {
	return cli.datas.Int(cccommon.BKHostIDField)
}

func (cli *host) GetInstName() string {
	return cli.datas.String(cccommon.BKHostIDField)
}

func (cli *host) GetValues() (types.MapStr, error) {
	return cli.datas, nil
}

func (cli *host) SetValue(key string, value interface{}) error {
	cli.datas.Set(key, value)
	return nil
}

func (cli *host) ResetAssociationValue() error {
	attrs, err := cli.target.Attributes()
	if nil != err {
		return err
	}

	for _, attrItem := range attrs {
		if attrItem.GetKey() {

			if model.FieldTypeSingleAsst == attrItem.GetType() {
				asstVals, err := cli.datas.MapStrArray(attrItem.GetID())
				if nil != err {
					return err
				}

				for _, val := range asstVals {

					valID, err := val.Int64(InstID)
					if nil != err {
						return err
					}
					cli.datas.Set(attrItem.GetID(), valID)
				}

				continue
			}

			if model.FieldTypeMultiAsst == attrItem.GetType() {
				asstVals, err := cli.datas.MapStrArray(attrItem.GetID())
				if nil != err {
					return err
				}

				condVals := make([]string, 0)
				for _, val := range asstVals {

					valID := val.String(InstID)
					if 0 != len(valID) {
						condVals = append(condVals, valID)
					}
				}

				cli.datas.Set(attrItem.GetID(), strings.Join(condVals, ","))
			}
		}
	}

	return nil
}

func (cli *host) search() ([]model.Attribute, []types.MapStr, error) {

	if err := cli.ResetAssociationValue(); nil != err {
		return nil, nil, err
	}

	attrs, err := cli.target.Attributes()
	if nil != err {
		return nil, nil, err
	}

	cond := common.CreateCondition()
	for _, attrItem := range attrs {
		if attrItem.GetKey() {

			attrVal, exists := cli.datas.Get(attrItem.GetID())
			if !exists {
				return nil, nil, errors.New("the key field(" + attrItem.GetID() + ") is not set")
			}
			cond.Field(attrItem.GetID()).Eq(attrVal)
		}
	}
	//log.Infof("the condition:%#v", cond.ToMapStr())
	// search by condition
	items, err := client.GetClient().CCV3().Host().SearchHost(cond)
	return attrs, items, err
}
func (cli *host) IsExists() (bool, error) {

	attrs, err := cli.target.Attributes()
	if nil != err {
		return false, err
	}

	cond := common.CreateCondition()

	for _, attrItem := range attrs {

		if !attrItem.GetKey() {
			continue
		}

		if !cli.datas.Exists(attrItem.GetID()) {
			continue
		}

		cond.Field(attrItem.GetID()).Eq(cli.datas[attrItem.GetID()])

	}

	items, err := client.GetClient().CCV3().Host().SearchHost(cond)
	if nil != err {
		return false, err
	}

	return 0 != len(items), nil
}
func (cli *host) Create() error {

	if exists, err := cli.IsExists(); nil != err {
		return err
	} else if exists {
		return nil
	}

	return cli.Save()

}
func (cli *host) Update() error {

	attrs, existItems, err := cli.search()
	if nil != err {
		return err
	}

	// clear the invalid data
	cli.datas.ForEach(func(key string, val interface{}) {
		for _, attrItem := range attrs {
			if attrItem.GetID() == key {
				return
			}
		}

		cli.datas.Remove(key)
	})

	for _, existItem := range existItems {

		hostID, err := existItem.Int64(HostID)
		if nil != err {
			return err
		}

		updateCond := common.CreateCondition()
		updateCond.Field(ModuleID).Eq(hostID)
		log.Infof("the exists:%s %d", string(cli.datas.ToJSON()), hostID)
		err = client.GetClient().CCV3().Host().UpdateHostBatch(cli.datas, strconv.Itoa(int(hostID)))
		if err != nil {
			log.Errorf("failed to update host, error info is %s", err.Error())
			return err
		}

	}

	return nil
}
func (cli *host) Save() error {

	if exists, err := cli.IsExists(); nil != err {
		return err
	} else if exists {
		return cli.Update()
	}

	return cli.Create()
}
