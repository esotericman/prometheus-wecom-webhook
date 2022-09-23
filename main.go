package main

/*
 * Copyright 2022 Flmelody
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/golang/glog"
)

//goland:noinspection SpellCheckingInspection
const (
	wecomHookUrl = "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=%s"
	port         = 6666
)

type KV map[string]string
type RData struct {
	Receiver string `json:"receiver"`
	Status   string `json:"status"`
	Alerts   Alerts `json:"alerts"`

	GroupLabels       KV `json:"groupLabels"`
	CommonLabels      KV `json:"commonLabels"`
	CommonAnnotations KV `json:"commonAnnotations"`

	ExternalURL string `json:"externalURL"`
}

// Alert holds one alert for notification templates.
type Alert struct {
	Status       string    `json:"status"`
	Labels       KV        `json:"labels"`
	Annotations  KV        `json:"annotations"`
	StartsAt     time.Time `json:"startsAt"`
	EndsAt       time.Time `json:"endsAt"`
	GeneratorURL string    `json:"generatorURL"`
	Fingerprint  string    `json:"fingerprint"`
}

// Alerts is a list of Alert objects.
type Alerts []Alert

// time formatter
func timeFormat(formatter string, ct time.Time) string {
	return ct.In(time.Local).Format(formatter)
}
func main() {
	http.HandleFunc("/webhook", func(rw http.ResponseWriter, req *http.Request) {
		// 反序列化请求数据
		decoder := json.NewDecoder(req.Body)
		var rd RData
		if err := decoder.Decode(&rd); err != nil {
			glog.Error(err)
			return
		}
		if rd.Alerts != nil && len(rd.Alerts) > 3 {
			newAlert := rd.Alerts[0:3]
			newAlert = append(newAlert, Alert{
				Status: "other",
			})
			rd.Alerts = newAlert
		}
		// 加载模板
		var tf = make(template.FuncMap)
		tf["timeFormat"] = timeFormat
		tmpl := template.Must(template.New("wecomhook.tmpl").Funcs(tf).ParseFiles("./template/wecomhook.tmpl"))
		var td bytes.Buffer
		if err := tmpl.Execute(&td, rd); err != nil {
			glog.Error(err)
			return
		}
		//goland:noinspection SpellCheckingInspection
		postBody, _ := json.Marshal(map[string]interface{}{
			"msgtype": "markdown",
			"markdown": map[string]interface{}{
				"content": td.String(),
			},
		})
		responseBody := bytes.NewBuffer(postBody)
		resp, err := http.Post(fmt.Sprintf(wecomHookUrl, os.Getenv("HOOK_KEY")), "application/json", responseBody)
		if err != nil {
			glog.Error(err)
			return
		}
		//goland:noinspection GoUnhandledErrorResult
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			if err != nil {
				glog.Error(err)
				rw.WriteHeader(http.StatusBadRequest)
				//goland:noinspection GoUnhandledErrorResult
				rw.Write([]byte(err.Error()))
			}
			glog.Error("Broken : ", string(body))
			rw.WriteHeader(http.StatusBadRequest)
			//goland:noinspection GoUnhandledErrorResult
			rw.Write(body)
		} else {
			glog.Info(fmt.Sprintf("Notify Success,Request->%s", postBody))
			rw.WriteHeader(http.StatusOK)
			//goland:noinspection GoUnhandledErrorResult
			rw.Write(body)
		}

	})
	fmt.Println(fmt.Sprintf("%s\tListen on Port %d", time.Now(), port))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		glog.Error(err)
		return
	}
}
