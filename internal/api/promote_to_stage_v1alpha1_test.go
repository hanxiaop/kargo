package api

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kargoapi "github.com/akuity/kargo/api/v1alpha1"
	fakekubeclient "github.com/akuity/kargo/internal/kubeclient/fake"
	svcv1alpha1 "github.com/akuity/kargo/pkg/api/service/v1alpha1"
)

func TestPromoteToStage(t *testing.T) {
	testCases := []struct {
		name       string
		req        *svcv1alpha1.PromoteToStageRequest
		server     *server
		assertions func(
			*testing.T,
			*fakekubeclient.EventRecorder,
			*connect.Response[svcv1alpha1.PromoteToStageResponse],
			error,
		)
	}{
		{
			name:   "input validation error",
			req:    &svcv1alpha1.PromoteToStageRequest{},
			server: &server{},
			assertions: func(
				t *testing.T,
				_ *fakekubeclient.EventRecorder,
				_ *connect.Response[svcv1alpha1.PromoteToStageResponse],
				err error,
			) {
				require.Error(t, err)
				var connErr *connect.Error
				require.True(t, errors.As(err, &connErr))
				require.Equal(t, connect.CodeInvalidArgument, connErr.Code())
			},
		},
		{
			name: "error validating project",
			req: &svcv1alpha1.PromoteToStageRequest{
				Project: "fake-project",
				Stage:   "fake-stage",
				Freight: "fake-freight",
			},
			server: &server{
				validateProjectExistsFn: func(context.Context, string) error {
					return errors.New("something went wrong")
				},
			},
			assertions: func(
				t *testing.T,
				_ *fakekubeclient.EventRecorder,
				_ *connect.Response[svcv1alpha1.PromoteToStageResponse],
				err error,
			) {
				require.Error(t, err)
				require.Equal(t, "something went wrong", err.Error())
			},
		},
		{
			name: "error getting Stage",
			req: &svcv1alpha1.PromoteToStageRequest{
				Project: "fake-project",
				Stage:   "fake-stage",
				Freight: "fake-freight",
			},
			server: &server{
				validateProjectExistsFn: func(context.Context, string) error {
					return nil
				},
				getStageFn: func(
					context.Context,
					client.Client,
					types.NamespacedName,
				) (*kargoapi.Stage, error) {
					return nil, errors.New("something went wrong")
				},
			},
			assertions: func(
				t *testing.T,
				_ *fakekubeclient.EventRecorder,
				_ *connect.Response[svcv1alpha1.PromoteToStageResponse],
				err error,
			) {
				require.Error(t, err)
				require.Equal(t, "get stage: something went wrong", err.Error())
			},
		},
		{
			name: "Stage not found",
			req: &svcv1alpha1.PromoteToStageRequest{
				Project: "fake-project",
				Stage:   "fake-stage",
				Freight: "fake-freight",
			},
			server: &server{
				validateProjectExistsFn: func(context.Context, string) error {
					return nil
				},
				getStageFn: func(
					context.Context,
					client.Client,
					types.NamespacedName,
				) (*kargoapi.Stage, error) {
					return nil, nil
				},
			},
			assertions: func(
				t *testing.T,
				_ *fakekubeclient.EventRecorder,
				_ *connect.Response[svcv1alpha1.PromoteToStageResponse],
				err error,
			) {
				require.Error(t, err)
				var connErr *connect.Error
				require.True(t, errors.As(err, &connErr))
				require.Equal(t, connect.CodeNotFound, connErr.Code())
				require.Contains(t, connErr.Message(), "Stage")
				require.Contains(t, connErr.Message(), "not found in namespace")
			},
		},
		{
			name: "error getting Freight",
			req: &svcv1alpha1.PromoteToStageRequest{
				Project: "fake-project",
				Stage:   "fake-stage",
				Freight: "fake-freight",
			},
			server: &server{
				validateProjectExistsFn: func(context.Context, string) error {
					return nil
				},
				getStageFn: func(
					context.Context,
					client.Client,
					types.NamespacedName,
				) (*kargoapi.Stage, error) {
					return &kargoapi.Stage{
						Spec: &kargoapi.StageSpec{
							Subscriptions: &kargoapi.Subscriptions{
								UpstreamStages: []kargoapi.StageSubscription{
									{
										Name: "fake-upstream-stage",
									},
								},
							},
						},
					}, nil
				},
				getFreightByNameOrAliasFn: func(
					context.Context,
					client.Client,
					string, string, string,
				) (*kargoapi.Freight, error) {
					return nil, errors.New("something went wrong")
				},
			},
			assertions: func(
				t *testing.T,
				_ *fakekubeclient.EventRecorder,
				_ *connect.Response[svcv1alpha1.PromoteToStageResponse],
				err error,
			) {
				require.Error(t, err)
				require.Equal(t, "get freight: something went wrong", err.Error())
			},
		},
		{
			name: "Freight not found",
			req: &svcv1alpha1.PromoteToStageRequest{
				Project: "fake-project",
				Stage:   "fake-stage",
				Freight: "fake-freight",
			},
			server: &server{
				validateProjectExistsFn: func(context.Context, string) error {
					return nil
				},
				getStageFn: func(
					context.Context,
					client.Client,
					types.NamespacedName,
				) (*kargoapi.Stage, error) {
					return &kargoapi.Stage{
						Spec: &kargoapi.StageSpec{
							Subscriptions: &kargoapi.Subscriptions{
								UpstreamStages: []kargoapi.StageSubscription{
									{
										Name: "fake-upstream-stage",
									},
								},
							},
						},
					}, nil
				},
				getFreightByNameOrAliasFn: func(
					context.Context,
					client.Client,
					string, string, string,
				) (*kargoapi.Freight, error) {
					return nil, nil
				},
			},
			assertions: func(
				t *testing.T,
				_ *fakekubeclient.EventRecorder,
				_ *connect.Response[svcv1alpha1.PromoteToStageResponse],
				err error,
			) {
				require.Error(t, err)
				var connErr *connect.Error
				require.True(t, errors.As(err, &connErr))
				require.Equal(t, connect.CodeNotFound, connErr.Code())
				require.Contains(t, connErr.Message(), "freight")
				require.Contains(t, connErr.Message(), "not found in namespace")
			},
		},
		{
			name: "Freight not available",
			req: &svcv1alpha1.PromoteToStageRequest{
				Project: "fake-project",
				Stage:   "fake-stage",
				Freight: "fake-freight",
			},
			server: &server{
				validateProjectExistsFn: func(context.Context, string) error {
					return nil
				},
				getStageFn: func(
					context.Context,
					client.Client,
					types.NamespacedName,
				) (*kargoapi.Stage, error) {
					return &kargoapi.Stage{
						Spec: &kargoapi.StageSpec{
							Subscriptions: &kargoapi.Subscriptions{
								UpstreamStages: []kargoapi.StageSubscription{
									{
										Name: "fake-upstream-stage",
									},
								},
							},
						},
					}, nil
				},
				getFreightByNameOrAliasFn: func(
					context.Context,
					client.Client,
					string, string, string,
				) (*kargoapi.Freight, error) {
					return &kargoapi.Freight{}, nil
				},
				isFreightAvailableFn: func(*kargoapi.Freight, string, []string) bool {
					return false
				},
			},
			assertions: func(
				t *testing.T,
				_ *fakekubeclient.EventRecorder,
				_ *connect.Response[svcv1alpha1.PromoteToStageResponse],
				err error,
			) {
				require.Error(t, err)
				var connErr *connect.Error
				require.True(t, errors.As(err, &connErr))
				require.Equal(t, connect.CodeInvalidArgument, connErr.Code())
				require.Contains(t, connErr.Message(), "Freight")
				require.Contains(t, connErr.Message(), "is not available to Stage")
			},
		},
		{
			name: "promoting not authorized",
			req: &svcv1alpha1.PromoteToStageRequest{
				Project: "fake-project",
				Stage:   "fake-stage",
				Freight: "fake-freight",
			},
			server: &server{
				validateProjectExistsFn: func(context.Context, string) error {
					return nil
				},
				getStageFn: func(
					context.Context,
					client.Client,
					types.NamespacedName,
				) (*kargoapi.Stage, error) {
					return &kargoapi.Stage{
						Spec: &kargoapi.StageSpec{
							Subscriptions: &kargoapi.Subscriptions{
								UpstreamStages: []kargoapi.StageSubscription{
									{
										Name: "fake-upstream-stage",
									},
								},
							},
						},
					}, nil
				},
				getFreightByNameOrAliasFn: func(
					context.Context,
					client.Client,
					string,
					string,
					string,
				) (*kargoapi.Freight, error) {
					return &kargoapi.Freight{}, nil
				},
				isFreightAvailableFn: func(*kargoapi.Freight, string, []string) bool {
					return true
				},
				authorizeFn: func(
					context.Context,
					string,
					schema.GroupVersionResource,
					string,
					client.ObjectKey,
				) error {
					return errors.New("not authorized")
				},
			},
			assertions: func(
				t *testing.T,
				_ *fakekubeclient.EventRecorder,
				_ *connect.Response[svcv1alpha1.PromoteToStageResponse],
				err error,
			) {
				require.Error(t, err)
				require.Equal(t, "not authorized", err.Error())
			},
		},
		{
			name: "error creating Promotion",
			req: &svcv1alpha1.PromoteToStageRequest{
				Project: "fake-project",
				Stage:   "fake-stage",
				Freight: "fake-freight",
			},
			server: &server{
				validateProjectExistsFn: func(context.Context, string) error {
					return nil
				},
				getStageFn: func(
					context.Context,
					client.Client,
					types.NamespacedName,
				) (*kargoapi.Stage, error) {
					return &kargoapi.Stage{
						Spec: &kargoapi.StageSpec{
							Subscriptions: &kargoapi.Subscriptions{
								UpstreamStages: []kargoapi.StageSubscription{
									{
										Name: "fake-upstream-stage",
									},
								},
							},
						},
					}, nil
				},
				getFreightByNameOrAliasFn: func(
					context.Context,
					client.Client,
					string, string, string,
				) (*kargoapi.Freight, error) {
					return &kargoapi.Freight{}, nil
				},
				isFreightAvailableFn: func(*kargoapi.Freight, string, []string) bool {
					return true
				},
				authorizeFn: func(
					context.Context,
					string,
					schema.GroupVersionResource,
					string,
					client.ObjectKey,
				) error {
					return nil
				},
				createPromotionFn: func(
					context.Context,
					client.Object,
					...client.CreateOption,
				) error {
					return errors.New("something went wrong")
				},
			},
			assertions: func(
				t *testing.T,
				_ *fakekubeclient.EventRecorder,
				_ *connect.Response[svcv1alpha1.PromoteToStageResponse],
				err error,
			) {
				require.Error(t, err)
				require.Equal(t, "create promotion: something went wrong", err.Error())
			},
		},
		{
			name: "success",
			req: &svcv1alpha1.PromoteToStageRequest{
				Project: "fake-project",
				Stage:   "fake-stage",
				Freight: "fake-freight",
			},
			server: &server{
				validateProjectExistsFn: func(context.Context, string) error {
					return nil
				},
				getStageFn: func(
					context.Context,
					client.Client,
					types.NamespacedName,
				) (*kargoapi.Stage, error) {
					return &kargoapi.Stage{
						Spec: &kargoapi.StageSpec{
							Subscriptions: &kargoapi.Subscriptions{
								UpstreamStages: []kargoapi.StageSubscription{
									{
										Name: "fake-upstream-stage",
									},
								},
							},
						},
					}, nil
				},
				getFreightByNameOrAliasFn: func(
					context.Context,
					client.Client,
					string, string, string,
				) (*kargoapi.Freight, error) {
					return &kargoapi.Freight{}, nil
				},
				isFreightAvailableFn: func(*kargoapi.Freight, string, []string) bool {
					return true
				},
				authorizeFn: func(
					context.Context,
					string,
					schema.GroupVersionResource,
					string,
					client.ObjectKey,
				) error {
					return nil
				},
				createPromotionFn: func(
					context.Context,
					client.Object,
					...client.CreateOption,
				) error {
					return nil
				},
			},
			assertions: func(
				t *testing.T,
				recorder *fakekubeclient.EventRecorder,
				res *connect.Response[svcv1alpha1.PromoteToStageResponse],
				err error,
			) {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.NotNil(t, res.Msg.GetPromotion())
				require.Len(t, recorder.Events, 1)
				event := <-recorder.Events
				require.Equal(t, corev1.EventTypeNormal, event.EventType)
				require.Equal(t, kargoapi.EventReasonPromotionCreated, event.Reason)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := fakekubeclient.NewEventRecorder(1)
			testCase.server.recorder = recorder
			res, err := testCase.server.PromoteToStage(
				context.Background(),
				connect.NewRequest(testCase.req),
			)
			testCase.assertions(t, recorder, res, err)
		})
	}
}
