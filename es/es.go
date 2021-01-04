/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 11:54:48
 * @LastEditTime: 2020-12-16 11:54:49
 * @LastEditors: Chen Long
 * @Reference:
 */

package es

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ChenLong-dev/gobase/config"
	"github.com/ChenLong-dev/gobase/etcd"
	"github.com/ChenLong-dev/gobase/mlog"
	"github.com/olivere/elastic"
)

var (
	Client *elastic.Client
)

//es操作为post请求服务器操作，直接使用Init对Client配置进行初始化
func init() {
	//判断config中es字段中key是否为空
	if config.Conf.Es.Key != "" {
		etcd.Connect()
		defer etcd.Close()
		//从etcd获取字段
		esConf, err := etcd.Get(config.Conf.Es.Key)
		if err != nil {
			mlog.Errorf("[ES] Get etcd_config err: %s \n", err)
		}
		mlog.Infof("[ES] config: %s \n", esConf)
		if err = json.Unmarshal(esConf, &config.Conf.Es); err != nil {
			mlog.Errorf("[ES] Parse etcd_config err: %s \n", err)
		}
	}
	if err := NewClient(); err != nil {
		return
	}
}

func NewClient() error {
	var err error
	Client, err = elastic.NewClient(
		elastic.SetURL(config.Conf.Es.Host),
		elastic.SetBasicAuth(config.Conf.Es.UserName, config.Conf.Es.Password))
	if err != nil {
		mlog.Error(err)
		return err
	}

	_, _, err = Client.Ping(config.Conf.Es.Host).Do(context.Background())
	if err != nil {
		mlog.Error(err)
		return err
	}

	_, err = Client.ElasticsearchVersion(config.Conf.Es.Host)
	if err != nil {
		mlog.Error(err)
		return err
	}

	return nil
}

// 创建 elasticSearch 的 Mapping
func InitMapping(esIndexName, esMapping string) error {
	var err error

	// Use the IndexExists service to check if a specified index exists.
	exists, err := Client.IndexExists(esIndexName).Do(context.Background())
	if err != nil {
		mlog.Errorf("IndexExists:%s", err.Error())
		return err
	}

	if !exists {
		mlog.Info("es index not exists: " + esIndexName)
		// Create a new index.
		createIndex, err := Client.CreateIndex(esIndexName).Body(esMapping).Do(context.Background())
		if err != nil {
			mlog.Error("CreateIndex: " + err.Error())
			return err
		}

		if !createIndex.Acknowledged {
			// Not acknowledged
			return errors.New("create index:" + esIndexName + ", not Ack nowledged")
		}
	}

	return nil
}

func Create(index string, body interface{}) error {
	_, err := Client.Index().
		Index(index).
		BodyJson(&body).
		Do(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func Delete(index, id string) error {
	_, err := Client.Delete().
		Index(index).
		Id(id).
		Do(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func Update(index, id string, body interface{}) error {
	_, err := Client.Update().
		Index(index).
		Id(id).
		Doc(&body).
		Do(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func Get(index, id string) (fields map[string]interface{}, err error) {
	res, e := Client.Get().
		Index(index).
		Id(id).
		Do(context.Background())
	if e != nil {
		return nil, e
	}

	if res.Found {
		fields = res.Fields
	}

	return
}

func List(index, id string, size, page int) (res *elastic.SearchResult, err error) {
	if size < 0 || page < 1 {
		return
	}

	res, err = Client.Search(index).
		Size(size).
		From((page - 1) * size).
		Do(context.Background())

	return
}

func Search(index string, query elastic.Query) (res *elastic.SearchResult, err error) {
	res, err = Client.Search(index).
		Query(query).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	return
}
