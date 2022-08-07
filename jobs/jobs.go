package jobs

import (
	"fmt"
	"github.com/xfyun/athena_init_config/utils"
	"sync"
	"time"
)

func DoInitConfigJob(c *utils.CenterService, done func()) {
	defer done()
	for {
		err := InitPush(c)
		if err != nil {
			time.Sleep(10 * time.Second)
		} else {
			break
		}
	}

}

func InitPush(c *utils.CenterService) error {
	loop := 0
	sok := false
	fmt.Println("waiting the config center ok...")
	for {
		if loop < 200 {
			if !sok {
				sok = c.IsAlive()
				if sok {
					break
				}
				time.Sleep(10 * time.Second)
				loop += 1
			} else {
				fmt.Println("Err execute because of service is not ok")
				break
			}
		}
	}
	fmt.Println("start pushing...")

	regions := []string{
		"test",
	}

	init_project := "test-project"
	init_cluster := "test-cluster"

	iservice := "webgate"
	version := "v1"
	var wfs = []string{
		"schema_s5ca83460.json", "app.toml", "xsf.toml",
	}
	for _, wf := range wfs {
		err := c.GetAndPushConfig(init_project, init_cluster, iservice, version, wf, regions)
		if err != nil {
			return err
		}
		fmt.Println("webgate config init ok...", wf)

	}

	iservice = "mmocr"
	cf := "aiges-remote.toml"
	err := c.GetAndPushConfig(init_project, init_cluster, iservice, version, cf, regions)
	if err != nil {
		return err
	}
	fmt.Println("mmocr config init ok...", cf)

	iservice = "loadbalance"
	cf = "lbv2.toml"
	version = "4.2.1"
	err = c.GetAndPushConfig(init_project, init_cluster, iservice, version, cf, regions)
	if err != nil {
		return err
	}
	fmt.Println("loadbalance config init ok...", cf)

	iservice = "atmos"
	version = "1.0.0"

	var cfs = []string{
		"xsfc.toml", "xsfs.toml", "atmos.cfg",
	}
	for _, cs := range cfs {
		err = c.GetAndPushConfig(init_project, init_cluster, iservice, version, cs, regions)
		if err != nil {
			return err
		}
		fmt.Println("atmos config init ok...", cs)
	}

	// update confing
	////err, resp := configService.DownLoadFile()
	////err, resp := configService.DeleteConfig()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func Execute(c *utils.CenterService) error {
	var wg sync.WaitGroup
	wg.Add(1)
	go DoInitConfigJob(c, wg.Done)
	wg.Wait()
	return nil
}
