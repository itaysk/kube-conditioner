/*
Copyright 2019 Itay Shakury.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package prometheus

import (
	"context"
	"time"

	prometheus "github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/model"
)

//PrometheusSource is a prometheus implementation of a datasource
type PrometheusSource struct {
	//ServerURL is the address where prometheus is running
	ServerURL string
	//Rule is the prometheus rule to evaluate
	Rule string
}

// NewPrometheusSource creates a new PrometheusSource
func NewPrometheusSource(serverURL string, rule string) *PrometheusSource {
	return &PrometheusSource{
		ServerURL: serverURL,
		Rule:      rule,
	}
}

//Evaluate is
func (p PrometheusSource) Evaluate() (bool, error) {
	config := prometheus.Config{Address: p.ServerURL, RoundTripper: prometheus.DefaultRoundTripper}
	client, err := prometheus.NewClient(config)
	if err != nil {
		log.Error(err, "error creating prometheus client")
	}
	httpAPI := v1.NewAPI(client)
	result, err := httpAPI.Query(context.TODO(), p.Rule, time.Now())
	if err != nil {
		log.Error(err, "error querying prometheus")
		return false, err //TODO: do I have to return false? maybe switch to return corev1.ConditionStatus instead of bool
	}
	v := []*model.Sample(result.(model.Vector))
	//in case of falsely query, there will be no series returnd
	if len(v) > 0 {
		return v[0].Value == 1, nil
	} else { //query succeeded and evaluates to false
		return false, nil
	}
}
