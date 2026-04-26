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

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
)

// ImageMetadataProvider resolves metadata from OCI images.
type ImageMetadataProvider interface {
	// GetImageLabels returns OCI image labels for the given image reference.
	GetImageLabels(ctx context.Context, image string) (map[string]string, error)
}

// resolveRepositoryURL extracts a source repository URL associated with a Pod.
//
// Resolution order:
//  1. Pod annotation: court.dev/repository
//  2. OCI image label: org.opencontainers.image.source
//
// Returns an empty string if no repository information can be resolved.
func resolveRepositoryURL(ctx context.Context, log logr.Logger, pod *v1.Pod, imgProvider ImageMetadataProvider) string {

	if pod == nil {
		return ""
	}

	// Eexplicit annotation (fast path)
	if repo, ok := pod.Annotations["court.dev/repository"]; ok && repo != "" {
		return repo
	}

	log.V(1).Info("repository not found in annotations, falling back to image metadata")

	if imgProvider == nil {
		log.V(1).Info("image metadata provider is not configured")
		return ""
	}

	// 2. OCI image metadata
	for _, c := range pod.Spec.Containers {

		labels, err := imgProvider.GetImageLabels(ctx, c.Image)
		if err != nil {
			log.Error(err, "failed to inspect image metadata", "image", c.Image)
			continue
		}

		if repo, ok := labels["org.opencontainers.image.source"]; ok && repo != "" {
			return repo
		}
	}

	log.V(1).Info("repository could not be resolved from annotations or image metadata")

	return ""
}
