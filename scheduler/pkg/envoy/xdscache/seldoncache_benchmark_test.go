/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package xdscache

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/sirupsen/logrus"
)

// Prevent compiler from optimising away benchmarks
var results []types.Resource

func benchmarkRouteContents(b *testing.B, numResources uint) {
	x := NewSeldonXDSCache(logrus.New(), nil)

	for n := 0; n < int(numResources); n++ {
		x.AddPipelineRoute(strconv.Itoa(n), strconv.Itoa(n), 100, false)
		x.AddRouteClusterTraffic(
			fmt.Sprintf("model-%d", n),
			fmt.Sprintf("model-%d", n),
			1,
			100,
			"http",
			"grpc",
			false,
			false,
		)
	}

	// Prevent compiler optimising away function calls
	var r []types.Resource
	for i := 0; i < b.N; i++ {
		r = x.RouteContents()
	}
	results = r
}

func BenchmarkRouteContents100(b *testing.B) { benchmarkRouteContents(b, 100) }
func BenchmarkRouteContents1K(b *testing.B)  { benchmarkRouteContents(b, 1_000) }
func BenchmarkRouteContents10K(b *testing.B) { benchmarkRouteContents(b, 10_000) }
