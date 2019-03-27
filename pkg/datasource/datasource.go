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

package datasource

//DataSource is a source of truth that can evaluate the state of a condition
type DataSource interface {
	Evaluate() (bool, error)
}

//DataSourceMock is a basic implementation of the DataSource interface
type DataSourceMock struct {
	// Result is the result to return
	Result bool
}

//Evaluate returns the value of Result in this mock
func (dsm DataSourceMock) Evaluate() (bool, error) {
	return dsm.Result, nil
}
