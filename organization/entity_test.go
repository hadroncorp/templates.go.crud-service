package organization_test

import (
	"testing"

	"github.com/hadroncorp/geck/security/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hadroncorp/service-template/organization"
)

func TestNew(t *testing.T) {
	ctx := identity.WithPrincipal(t.Context(), identity.NewBasicPrincipal("foo"))
	tests := []struct {
		name   string
		inID   string
		inName string
	}{
		{
			name:   "valid",
			inID:   "1",
			inName: "acme-corp",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := organization.New(ctx, tt.inID, tt.inName)
			assert.Equal(t, tt.inID, out.ID())
			assert.Equal(t, tt.inName, out.Name())
			assert.NotZero(t, out.CreateTime())
			assert.NotZero(t, out.CreateBy())

			// assert event generation
			events := out.PullEvents()
			require.Len(t, events, 1)
			assert.Equal(t, "hadron.iam.organization.created", events[0].Topic().String())
			payload, err := events[0].Bytes()
			assert.NoError(t, err)
			assert.NotZero(t, len(payload))
		})
	}
}

func TestOrganization_Update(t *testing.T) {
	ctx := identity.WithPrincipal(t.Context(), identity.NewBasicPrincipal("foo"))
	tests := []struct {
		name       string
		inOpts     []organization.UpdateOption
		expName    string
		expUpdated bool
		expEvents  int
	}{
		{
			name: "valid",
			inOpts: []organization.UpdateOption{
				organization.WithUpdatedName("acme-corp-2"),
			},
			expName:    "acme-corp-2",
			expUpdated: true,
			expEvents:  2,
		},
		{
			name:       "no-op",
			inOpts:     nil,
			expName:    "acme-corp",
			expUpdated: false,
			expEvents:  1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			org := organization.New(ctx, "1", "acme-corp")
			updated := org.Update(ctx, tt.inOpts...)
			assert.Equal(t, tt.expUpdated, updated)
			assert.Equal(t, tt.expName, org.Name())
			assert.NotZero(t, org.LastUpdateTime())
			assert.NotZero(t, org.LastUpdateBy())

			// assert event generation
			events := org.PullEvents()
			require.Len(t, events, tt.expEvents)
			assert.Equal(t, "hadron.iam.organization.created", events[0].Topic().String())

			if tt.expEvents == 1 {
				return
			}
			assert.Equal(t, "hadron.iam.organization.updated", events[1].Topic().String())
			payload, err := events[1].Bytes()
			assert.NoError(t, err)
			assert.NotZero(t, len(payload))
		})
	}
}

func TestOrganization_Delete(t *testing.T) {
	ctx := identity.WithPrincipal(t.Context(), identity.NewBasicPrincipal("foo"))
	org := organization.New(ctx, "1", "acme-corp")
	org.Delete(ctx)
	assert.True(t, org.IsDeleted())
	assert.NotZero(t, org.LastUpdateTime())
	assert.NotZero(t, org.LastUpdateBy())

	// assert event generation
	events := org.PullEvents()
	require.Len(t, events, 2)
	assert.Equal(t, "hadron.iam.organization.created", events[0].Topic().String())
	assert.Equal(t, "hadron.iam.organization.deleted", events[1].Topic().String())
	payload, err := events[1].Bytes()
	assert.NoError(t, err)
	assert.NotZero(t, len(payload))
}
