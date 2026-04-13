/*
Copyright 2026 Pedro Cozinheiro.

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

package officer

import (
	"context"

	"github.com/ghdrope/court/internal/incident"
	"github.com/ghdrope/court/internal/router"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pb "github.com/ghdrope/court/proto/incident"
	ctrl "sigs.k8s.io/controller-runtime"
)

// PodReconciler reconciles Pod resources and produces IncidentReports,
// for workloads that match known failure conditions.
type PodReconciler struct {
	client.Client
	Log logr.Logger

	API router.IncidentSender
}

// Reconcile evaluates the current state of a Pod and determines whether
// it should generate an IncidentReport.
//
// A report is created when:
//   - the Pod is in a failed/unknown phase OR
//   - container-level issues are detected by the inspector
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	logger := r.Log.WithValues(
		"namespace", req.Namespace,
		"name", req.Name,
	)

	var pod v1.Pod

	// Fetch latest Pod state
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Inspect container-level issues such as CrashLoopBackOff or OOMKilled.
	containerIssues := DetectContainerIssues(&pod)

	isProblem :=
		pod.Status.Phase == v1.PodFailed ||
			pod.Status.Phase == v1.PodUnknown ||
			len(containerIssues) > 0

	if !isProblem {
		return ctrl.Result{}, nil
	}

	// Build a structured incident report from the Pod state.
	report, err := incident.BuildFromPod(
		&pod,
		nil,
		containerIssues,
	)
	if err != nil {
		logger.Error(err, "failed to build incident report")
		return ctrl.Result{}, err
	}

	// Map domain report to protobuf format
	pbReport := &pb.IncidentReport{
		Id:        report.ID,
		Namespace: report.Namespace,
	}

	for _, ci := range report.ContainerIssues {
		pbReport.ContainerIssues = append(pbReport.ContainerIssues, &pb.ContainerIssue{
			Container: ci.Container,
			Reason:    ci.Reason,
		})
	}

	logger.Info("sending incident", "id", report.ID)

	if err := r.API.Send(ctx, pbReport); err != nil {
		logger.Error(err, "failed to send incident")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
