/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package scheduler

import (
	"context"
	"fmt"
	"io"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
)

func (s *SchedulerClient) ServerNotify(ctx context.Context, grpcClient scheduler.SchedulerClient, servers []v1alpha1.Server, isFirstSync bool) error {
	logger := s.logger.WithName("NotifyServer")
	if len(servers) == 0 {
		return nil
	}

	if grpcClient == nil {
		// we assume that all servers are in the same namespace
		namespace := servers[0].Namespace
		conn, err := s.getConnection(namespace)
		if err != nil {
			return err
		}
		grpcClient = scheduler.NewSchedulerClient(conn)
	}

	var requests []*scheduler.ServerNotify
	for _, server := range servers {
		var replicas int32
		if !server.ObjectMeta.DeletionTimestamp.IsZero() {
			replicas = 0
		} else if server.Spec.Replicas != nil {
			replicas = *server.Spec.Replicas
		} else {
			replicas = 1
		}

		logger.Info("Notify server", "name", server.GetName(), "namespace", server.GetNamespace(), "replicas", replicas)
		requests = append(requests, &scheduler.ServerNotify{
			Name:             server.GetName(),
			ExpectedReplicas: replicas,
			KubernetesMeta: &scheduler.KubernetesMeta{
				Namespace:  server.GetNamespace(),
				Generation: server.GetGeneration(),
			},
		})
	}
	request := &scheduler.ServerNotifyRequest{
		Servers:     requests,
		IsFirstSync: isFirstSync,
	}
	_, err := grpcClient.ServerNotify(
		ctx,
		request,
		grpc_retry.WithMax(SchedulerConnectMaxRetries),
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(SchedulerConnectBackoffScalar)),
	)
	if err != nil {
		logger.Error(err, "Failed to send notify server to scheduler")
		return err
	}
	return nil
}

// note: namespace is not used in this function
func (s *SchedulerClient) SubscribeServerEvents(ctx context.Context, grpcClient scheduler.SchedulerClient, namespace string) error {
	logger := s.logger.WithName("SubscribeServerEvents")

	stream, err := grpcClient.SubscribeServerStatus(
		ctx,
		&scheduler.ServerSubscriptionRequest{SubscriberName: "seldon manager"},
		grpc_retry.WithMax(SchedulerConnectMaxRetries),
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(SchedulerConnectBackoffScalar)),
	)
	if err != nil {
		return err
	}

	// on new reconnects we send a list of servers to the scheduler
	go handleRegisteredServers(ctx, namespace, s, grpcClient)

	for {
		event, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Error(err, "event recv failed")
			return err
		}

		logger.Info("Received event", "server", event.ServerName, "raw", event)
		kubernetesMeta := event.GetKubernetesMeta()
		if kubernetesMeta == nil {
			kubernetesMeta = &scheduler.KubernetesMeta{
				Namespace: namespace,
			}
		}
		server := &v1alpha1.Server{}
		err = s.Get(ctx, client.ObjectKey{Name: event.ServerName, Namespace: kubernetesMeta.GetNamespace()}, server)
		if err != nil {
			logger.Error(err, "Failed to get server", "name", event.ServerName, "namespace", kubernetesMeta.GetNamespace())
			continue
		}

		// Try to update status
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			server := &v1alpha1.Server{}
			err = s.Get(ctx, client.ObjectKey{Name: event.ServerName, Namespace: kubernetesMeta.GetNamespace()}, server)
			if err != nil {
				return err
			}
			if event.GetKubernetesMeta() != nil &&
				event.GetKubernetesMeta().Generation != server.Generation {
				logger.Info("Ignoring event for old generation", "currentGeneration", server.Generation, "eventGeneration", event.GetKubernetesMeta().Generation, "server", event.ServerName)
				return nil
			}
			// Handle status update
			server.Status.LoadedModelReplicas = event.NumLoadedModelReplicas
			server.Status.ReplicasConnectedToControlPlane = event.AvailableReplicas
			atLeastMinConnected :=
				server.Spec.MinReplicas != nil &&
					event.AvailableReplicas >= *server.Spec.MinReplicas
			allConnected :=
				event.AvailableReplicas >= server.Status.Replicas
			replicaConnectionStatusType := v1alpha1.SomeReplicasConnected
			replicaConnectionStatusMsg := fmt.Sprintf("%d/%d replicas available", event.AvailableReplicas, event.ExpectedReplicas)
			if allConnected {
				if server.Status.Replicas != 0 {
					replicaConnectionStatusType = v1alpha1.AllReplicasConnected
					replicaConnectionStatusMsg = "All replicas connected to control plane"
				} else {
					replicaConnectionStatusType = v1alpha1.NoReplicasConnected
					replicaConnectionStatusMsg = "This server has no replicas"
				}
			}
			if server.Status.Replicas != 0 {
				server.Status.CreateAndSetCondition(
					v1alpha1.ControlPlaneConnectionsReady,
					atLeastMinConnected || allConnected,
					string(replicaConnectionStatusType),
					replicaConnectionStatusMsg)
			} else {
				server.Status.CreateAndSetCondition(
					v1alpha1.ControlPlaneConnectionsReady,
					true,
					string(replicaConnectionStatusType),
					replicaConnectionStatusMsg)
			}
			logger.Info("Setting server status", "server", server)
			return s.updateServerStatus(server)
		})
		if retryErr != nil {
			logger.Error(err, "Failed to update status", "model", event.ServerName)
		}

	}
	return nil
}

func (s *SchedulerClient) updateServerStatus(server *v1alpha1.Server) error {
	if err := s.Status().Update(context.TODO(), server); err != nil {
		s.recorder.Eventf(server, v1.EventTypeWarning, "UpdateFailed",
			"Failed to update status for Server %q: %v", server.Name, err)
		return err
	}
	return nil
}
