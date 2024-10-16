/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package server

import (
	"context"
	"testing"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	logrtest "github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	auth "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	client2 "sigs.k8s.io/controller-runtime/pkg/client"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/controllers/reconcilers/common"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
	testing2 "github.com/seldonio/seldon-core/operator/v2/pkg/utils/testing"
)

func TestServiceReconcile(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name             string
		serviceConfig    mlopsv1alpha1.ServiceConfig
		runtime          *mlopsv1alpha1.OverrideSpec
		overrides        map[string]*mlopsv1alpha1.OverrideSpec
		expectedSvcNames []string
		expectedSvcType  map[string]v1.ServiceType
	}
	tests := []test{
		{
			name: "normal services",
			serviceConfig: mlopsv1alpha1.ServiceConfig{
				GrpcServicePrefix: "",
			},
			overrides:        map[string]*mlopsv1alpha1.OverrideSpec{},
			expectedSvcNames: []string{SeldonMeshSVCName, mlopsv1alpha1.SchedulerName, mlopsv1alpha1.PipelineGatewayName},
			expectedSvcType: map[string]v1.ServiceType{
				SeldonMeshSVCName:                 v1.ServiceTypeLoadBalancer,
				mlopsv1alpha1.PipelineGatewayName: "",
			},
		},
		{
			name: "prefix services",
			serviceConfig: mlopsv1alpha1.ServiceConfig{
				GrpcServicePrefix: "grpc-",
			},
			overrides:        map[string]*mlopsv1alpha1.OverrideSpec{},
			expectedSvcNames: []string{SeldonMeshSVCName, mlopsv1alpha1.SchedulerName, mlopsv1alpha1.PipelineGatewayName},
			expectedSvcType: map[string]v1.ServiceType{
				SeldonMeshSVCName:                 v1.ServiceTypeLoadBalancer,
				mlopsv1alpha1.PipelineGatewayName: "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := logrtest.New(t)
			var client client2.Client
			scheme := runtime.NewScheme()
			err := mlopsv1alpha1.AddToScheme(scheme)
			g.Expect(err).To(BeNil())
			err = v1.AddToScheme(scheme)
			g.Expect(err).To(BeNil())
			err = auth.AddToScheme(scheme)
			g.Expect(err).To(BeNil())
			err = appsv1.AddToScheme(scheme)
			g.Expect(err).To(BeNil())
			g.Expect(err).To(BeNil())
			annotator := patch.NewAnnotator(constants.LastAppliedConfig)
			meta := metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			}
			client = testing2.NewFakeClient(scheme)
			sr := NewComponentServiceReconciler(
				common.ReconcilerConfig{Ctx: context.TODO(), Logger: logger, Client: client},
				meta,
				test.serviceConfig,
				test.overrides,
				annotator)
			g.Expect(err).To(BeNil())
			err = sr.Reconcile()

			g.Expect(err).To(BeNil())
			for _, svcName := range test.expectedSvcNames {
				svc := &v1.Service{}
				err := client.Get(context.TODO(), types.NamespacedName{
					Name:      svcName,
					Namespace: meta.GetNamespace(),
				}, svc)
				g.Expect(err).To(BeNil())
				if test.expectedSvcType != nil {
					if svcType, ok := test.expectedSvcType[svcName]; ok {
						g.Expect(svc.Spec.Type).To(Equal(svcType))
					}
				}
			}
		})
	}
}
