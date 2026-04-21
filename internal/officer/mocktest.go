package officer

import (
	"context"

	"github.com/ghdrope/court/internal/incident"
	"github.com/ghdrope/court/internal/suit"
)

type fakeIncidentRepo struct {
	insertErr error
	called    bool
}

func (f *fakeIncidentRepo) Insert(
	ctx context.Context,
	r *incident.IncidentReport,
) error {
	f.called = true
	return f.insertErr
}

type fakeSuitRepo struct {
	list []suit.Suit
}

func (f *fakeSuitRepo) ListOpen(ctx context.Context) ([]suit.Suit, error) {
	return f.list, nil
}

func (f *fakeSuitRepo) Close(ctx context.Context, id string) error {
	return nil
}
