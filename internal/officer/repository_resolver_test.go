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
	"testing"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type fakeImageProvider struct {
	labels map[string]string
	err    error
}

func (f *fakeImageProvider) GetImageLabels(ctx context.Context, image string) (map[string]string, error) {
	return f.labels, f.err
}

// TestResolveRepositoryURL verifies repository resolution priority:
// 1. Pod annotation (highest priority)
// 2. OCI image label fallback
// 3. Empty result when no metadata is available
func TestResolveRepositoryURL(t *testing.T) {

	zapLog, _ := zap.NewDevelopment()
	log := zapr.NewLogger(zapLog)

	tests := []struct {
		name       string
		pod        *v1.Pod
		provider   ImageMetadataProvider
		expectRepo string
	}{
		{
			name: "annotation wins",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"court.dev/repository": "https://github.com/example/repo",
					},
				},
			},
			expectRepo: "https://github.com/example/repo",
		},
		{
			name: "image label fallback",
			pod: &v1.Pod{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{Image: "nginx:latest"},
					},
				},
			},
			provider: &fakeImageProvider{
				labels: map[string]string{
					"org.opencontainers.image.source": "https://github.com/image/repo",
				},
			},
			expectRepo: "https://github.com/image/repo",
		},
		{
			name: "no metadata returns empty",
			pod: &v1.Pod{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{Image: "nginx:latest"},
					},
				},
			},
			provider:   &fakeImageProvider{labels: map[string]string{}},
			expectRepo: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := resolveRepositoryURL(context.TODO(), log, tt.pod, tt.provider)

			if got != tt.expectRepo {
				t.Errorf("expected %q, got %q", tt.expectRepo, got)
			}
		})
	}
}
