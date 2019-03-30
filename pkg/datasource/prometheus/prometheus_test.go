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
	stdlog "log"
	"os"
	"testing"

	"github.com/onsi/gomega"
)

const PrometheusEnv = "TEST_PROMETHEUS_SERVER"

var PrometheusURL string
var PrometheusOK bool

func TestMain(m *testing.M) {
	if PrometheusURL, PrometheusOK = os.LookupEnv(PrometheusEnv); !PrometheusOK {
		stdlog.Printf("missing environment variable %s", PrometheusEnv)
	}

	code := m.Run()
	os.Exit(code)
}

func TestEvaluate(t *testing.T) {
	if !PrometheusOK {
		t.Skip("prometheus server not setup, skipping")
	}

	g := gomega.NewGomegaWithT(t)
	prom := NewPrometheusSource(PrometheusURL, "vector(1)>0")
	res, err := prom.Evaluate()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(res).To(gomega.BeTrue())
	prom = NewPrometheusSource(PrometheusURL, "vector(1)<0")
	res, err = prom.Evaluate()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(res).To(gomega.BeFalse())
}
