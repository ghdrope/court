package officer

import (
	"context"
	"fmt"

	"github.com/ghdrope/court/internal/incident"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// RecoverOpenSuits performs boot-time reconciliation of open suits.
//
// It ensures that after a crash or restart, Officer re-evaluates
// the real cluster state and corrects outdated incident assumptions.
func (s *Service) RecoverOpenSuits(
	ctx context.Context,
	cluster string,
	kube kubernetes.Interface,
) error {

	logger := s.Log.With(
		zap.String("phase", "recovery"),
		zap.String("cluster", cluster),
	)

	logger.Info("starting recovery of open suits")

	suits, err := s.SuitRepo.ListOpen(ctx)
	if err != nil {
		logger.Error("failed to fetch open suits", zap.Error(err))
		return fmt.Errorf("fetch open suits: %w", err)
	}

	logger.Info("open suits loaded", zap.Int("count", len(suits)))

	var (
		closedCount int
		keptCount   int
		errorCount  int
	)

	for _, suit := range suits {

		podName, ns, err := incident.ParseIncidentID(suit.IncidentID)
		if err != nil {
			logger.Warn("invalid incident id format",
				zap.String("incident_id", suit.IncidentID),
				zap.Error(err),
			)
			errorCount++
			continue
		}

		k8sPod, err := kube.CoreV1().
			Pods(ns).
			Get(ctx, podName, metav1.GetOptions{})

		// Case 1: Pod no longer exists -> close suit
		if err != nil {
			logger.Info("pod not found, closing suit",
				zap.String("suit_id", suit.ID),
				zap.String("pod", podName),
				zap.String("namespace", ns),
			)

			if closeErr := s.SuitRepo.Close(ctx, suit.ID); closeErr != nil {
				logger.Error("failed to close suit",
					zap.String("suit_id", suit.ID),
					zap.Error(closeErr),
				)
				errorCount++
			} else {
				closedCount++
			}

			continue
		}

		// Case 2: Pod is healthy -> self-healed -> close suit
		if isPodHealthy(k8sPod) {
			logger.Info("pod recovered, closing suit",
				zap.String("suit_id", suit.ID),
				zap.String("pod", podName),
			)

			if closeErr := s.SuitRepo.Close(ctx, suit.ID); closeErr != nil {
				logger.Error("failed to close suit",
					zap.String("suit_id", suit.ID),
					zap.Error(closeErr),
				)
				errorCount++
			} else {
				closedCount++
			}

			continue
		}

		// Case 3: still failing -> keep open
		logger.Debug("suit still active",
			zap.String("suit_id", suit.ID),
			zap.String("pod", podName),
		)

		keptCount++
	}

	logger.Info("recovery completed",
		zap.Int("closed", closedCount),
		zap.Int("kept", keptCount),
		zap.Int("errors", errorCount),
	)

	return nil
}

// isPodHealthy checks whether a pod is actually healthy,
// not just "Running".
func isPodHealthy(pod *v1.Pod) bool {

	if pod.Status.Phase != v1.PodRunning {
		return false
	}

	for _, c := range pod.Status.Conditions {
		if c.Type == v1.PodReady && c.Status == v1.ConditionTrue {
			return true
		}
	}

	return false
}
