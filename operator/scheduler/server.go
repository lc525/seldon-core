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

	var scalingSpec *v1alpha1.ValidatedScalingSpec
	if !server.ObjectMeta.DeletionTimestamp.IsZero() {
		scalingSpec = &v1alpha1.ValidatedScalingSpec{
			Replicas:    0,
			MinReplicas: 0,
			MaxReplicas: 0,
		}
	} else {
		scalingSpec, err = v1alpha1.GetValidatedScalingSpec(server.Spec.Replicas, server.Spec.MinReplicas, server.Spec.MaxReplicas)
		if err != nil {
			return err
		}
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
			MinReplicas:			scalingSpec.MinReplicas,
			MaxReplicas:			scalingSpec.MaxReplicas,
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

	// on new reconnects we send a list of servers to the schedule
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

		logger.Info("Received event", "server", event.ServerName)
		if event.GetKubernetesMeta() == nil {
			logger.Info("Received server event with no k8s metadata so ignoring", "server", event.ServerName)
			continue
		}
		server := &v1alpha1.Server{}
		err = s.Get(ctx, client.ObjectKey{Name: event.ServerName, Namespace: event.GetKubernetesMeta().GetNamespace()}, server)
		if err != nil {
			logger.Error(err, "Failed to get server", "name", event.ServerName, "namespace", event.GetKubernetesMeta().GetNamespace())
			continue
		}

		// Try to update status
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			server := &v1alpha1.Server{}
			err = s.Get(ctx, client.ObjectKey{Name: event.ServerName, Namespace: event.GetKubernetesMeta().GetNamespace()}, server)
			if err != nil {
				return err
			}
			if event.GetKubernetesMeta().Generation != server.Generation {
				logger.Info("Ignoring event for old generation", "currentGeneration", server.Generation, "eventGeneration", event.GetKubernetesMeta().Generation, "server", event.ServerName)
				return nil
			}

			// The types of updates we may get from the scheduler are:
			// 1. Status updates
			// 2. Requests for changing the number of server replicas
			// 3. Updates containing non-authoritative replica info, because the scheduler is in a
			// discovery phase (just starting up, after a restart)
			//
			// At the moment, the scheduler doesn't send multiple types of updates in a single event;
			switch event.GetType()	{
			case scheduler.ServerStatusResponse_StatusUpdate:
				return s.applyStatusUpdates(ctx, server, event)
			case scheduler.ServerStatusResponse_ScalingRequest:
				if event.ExpectedReplicas != event.AvailableReplicas {
					return s.applyReplicaUpdates(ctx, server, event)
				} else {
					return nil
				}
			case scheduler.ServerStatusResponse_NonAuthoritativeReplicaInfo:
				// skip updating replica info, only update status
				return s.updateServerStatus(server)
			default: // we ignore unknown event types
				return nil
			}
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

// when need to notify the scheduler about existing Server configuration
func handleRegisteredServers(
	ctx context.Context, namespace string, s *SchedulerClient, grpcClient scheduler.SchedulerClient) {
	serverList := &v1alpha1.ServerList{}
	// Get all servers in the namespace
	err := s.List(
		ctx,
		serverList,
		client.InNamespace(namespace),
	)
	if err != nil {
		return
	}

	for _, server := range serverList.Items {
		// servers that are not in the process of being deleted has DeletionTimestamp as zero
		if server.ObjectMeta.DeletionTimestamp.IsZero() {
			s.logger.V(1).Info("Calling NotifyServer (on reconnect)", "server", server.Name)
			if err := s.ServerNotify(ctx, &server); err != nil {
				s.logger.Error(err, "Failed to notify scheduler about initial Server parameters", "server", server.Name)
			} else {
				s.logger.V(1).Info("Load model called successfully", "server", server.Name)
			}
		} else {
			s.logger.V(1).Info("Server being deleted, not notifying", "server", server.Name)
		}
	}
}
