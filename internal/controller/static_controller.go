package controller

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	staticv1alpha1 "github.com/antonjah/static/pkg/apis/static/v1alpha1"
)

const (
	// Service account names
	StaticServiceAccount   = "static-service"
	OperatorServiceAccount = "static-operator"
)

type StaticReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *StaticReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var static staticv1alpha1.Static
	if err := r.Get(ctx, req.NamespacedName, &static); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "unable to fetch Static")
		return ctrl.Result{}, err
	}

	// Resource is being deleted - owner references handle cleanup automatically
	if !static.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	if err := r.reconcileDeployment(ctx, &static); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileService(ctx, &static); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, r.updateStatus(ctx, &static)
}

func (r *StaticReconciler) reconcileDeployment(ctx context.Context, static *staticv1alpha1.Static) error {
	logger := log.FromContext(ctx)

	replicas := int32(1)
	if static.Spec.Replicas != nil {
		replicas = *static.Spec.Replicas
	}

	image := "antonjah/static:latest"
	if static.Spec.Image != "" {
		image = static.Spec.Image
	}

	logLevel := "info"
	if static.Spec.LogLevel != "" {
		logLevel = static.Spec.LogLevel
	}

	labels := map[string]string{
		"app":                          "static",
		"app.kubernetes.io/name":       "static",
		"app.kubernetes.io/instance":   static.Name,
		"app.kubernetes.io/managed-by": "static-operator",
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      static.Name,
			Namespace: static.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, deployment, func() error {
		deployment.Labels = labels
		deployment.Spec.Replicas = &replicas
		deployment.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: labels,
		}

		deployment.Spec.Template = corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:            "static",
						Image:           image,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Ports: []corev1.ContainerPort{
							{
								Name:          "http",
								ContainerPort: 8080,
								Protocol:      corev1.ProtocolTCP,
							},
						},
						Env: []corev1.EnvVar{
							{
								Name:  "LOG_LEVEL",
								Value: logLevel,
							},
							{
								Name:  "IN_CLUSTER",
								Value: "true",
							},
							{
								Name: "NAMESPACE",
								ValueFrom: &corev1.EnvVarSource{
									FieldRef: &corev1.ObjectFieldSelector{
										FieldPath: "metadata.namespace",
									},
								},
							},
						},
					},
				},
			},
		}

		if static.Spec.TLS != nil && static.Spec.TLS.Enabled {
			// Validate: cannot specify both SecretName and file paths
			if static.Spec.TLS.SecretName != "" && (static.Spec.TLS.Certificate != "" || static.Spec.TLS.Key != "" || static.Spec.TLS.CA != "") {
				return fmt.Errorf("cannot specify both secretName and file paths (certificate/key/ca) in TLS config")
			}

			deployment.Spec.Template.Spec.Containers[0].Env = append(
				deployment.Spec.Template.Spec.Containers[0].Env,
				corev1.EnvVar{Name: "TLS_ENABLED", Value: "true"},
			)

			if static.Spec.TLS.SecretName != "" {
				deployment.Spec.Template.Spec.Containers[0].Env = append(
					deployment.Spec.Template.Spec.Containers[0].Env,
					corev1.EnvVar{Name: "TLS_CERTIFICATE", Value: "/tls/tls.crt"},
					corev1.EnvVar{Name: "TLS_KEY", Value: "/tls/tls.key"},
				)
				deployment.Spec.Template.Spec.Containers[0].VolumeMounts = append(
					deployment.Spec.Template.Spec.Containers[0].VolumeMounts,
					corev1.VolumeMount{
						Name:      "tls",
						MountPath: "/tls",
						ReadOnly:  true,
					},
				)
				deployment.Spec.Template.Spec.Volumes = append(
					deployment.Spec.Template.Spec.Volumes,
					corev1.Volume{
						Name: "tls",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: static.Spec.TLS.SecretName,
							},
						},
					},
				)

				if static.Spec.TLS.CA != "" {
					deployment.Spec.Template.Spec.Containers[0].Env = append(
						deployment.Spec.Template.Spec.Containers[0].Env,
						corev1.EnvVar{Name: "TLS_CA", Value: "/tls/ca.crt"},
					)
				}
			} else {
				if static.Spec.TLS.Certificate != "" {
					deployment.Spec.Template.Spec.Containers[0].Env = append(
						deployment.Spec.Template.Spec.Containers[0].Env,
						corev1.EnvVar{Name: "TLS_CERTIFICATE", Value: static.Spec.TLS.Certificate},
					)
				}
				if static.Spec.TLS.Key != "" {
					deployment.Spec.Template.Spec.Containers[0].Env = append(
						deployment.Spec.Template.Spec.Containers[0].Env,
						corev1.EnvVar{Name: "TLS_KEY", Value: static.Spec.TLS.Key},
					)
				}
				if static.Spec.TLS.CA != "" {
					deployment.Spec.Template.Spec.Containers[0].Env = append(
						deployment.Spec.Template.Spec.Containers[0].Env,
						corev1.EnvVar{Name: "TLS_CA", Value: static.Spec.TLS.CA},
					)
				}
			}

			if static.Spec.TLS.VerifyClient {
				deployment.Spec.Template.Spec.Containers[0].Env = append(
					deployment.Spec.Template.Spec.Containers[0].Env,
					corev1.EnvVar{Name: "TLS_VERIFY_CLIENT", Value: "true"},
				)
			}
		}

		if static.Spec.Resources != nil {
			deployment.Spec.Template.Spec.Containers[0].Resources = *static.Spec.Resources
		} else {
			deployment.Spec.Template.Spec.Containers[0].Resources = corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("200m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("50m"),
					corev1.ResourceMemory: resource.MustParse("64Mi"),
				},
			}
		}

		return controllerutil.SetControllerReference(static, deployment, r.Scheme)
	})

	if err != nil {
		logger.Error(err, "failed to reconcile Deployment")
		return err
	}

	logger.Info("Deployment reconciled", "name", deployment.Name)
	return nil
}

func (r *StaticReconciler) reconcileService(ctx context.Context, static *staticv1alpha1.Static) error {
	logger := log.FromContext(ctx)

	labels := map[string]string{
		"app":                          "static",
		"app.kubernetes.io/name":       "static",
		"app.kubernetes.io/instance":   static.Name,
		"app.kubernetes.io/managed-by": "static-operator",
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      static.Name,
			Namespace: static.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, service, func() error {
		service.Labels = labels
		service.Spec.Selector = labels
		service.Spec.Type = corev1.ServiceTypeClusterIP
		service.Spec.Ports = []corev1.ServicePort{
			{
				Name:       "http",
				Port:       80,
				TargetPort: intstr.FromInt(8080),
				Protocol:   corev1.ProtocolTCP,
			},
		}

		return controllerutil.SetControllerReference(static, service, r.Scheme)
	})

	if err != nil {
		logger.Error(err, "failed to reconcile Service")
		return err
	}

	logger.Info("Service reconciled", "name", service.Name)
	return nil
}

func (r *StaticReconciler) updateStatus(ctx context.Context, static *staticv1alpha1.Static) error {
	deployment := &appsv1.Deployment{}
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: static.Namespace,
		Name:      static.Name,
	}, deployment); err != nil {
		if errors.IsNotFound(err) {
			static.Status.Ready = false
			static.Status.Message = "Deployment not found"
			static.Status.Replicas = 0
		} else {
			return err
		}
	} else {
		static.Status.Replicas = deployment.Status.ReadyReplicas
		static.Status.Ready = deployment.Status.ReadyReplicas > 0
		if static.Status.Ready {
			static.Status.Message = fmt.Sprintf("%d/%d replicas ready",
				deployment.Status.ReadyReplicas,
				deployment.Status.Replicas)
		} else {
			static.Status.Message = "Waiting for replicas to be ready"
		}
	}

	return r.Status().Update(ctx, static)
}

func (r *StaticReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&staticv1alpha1.Static{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
