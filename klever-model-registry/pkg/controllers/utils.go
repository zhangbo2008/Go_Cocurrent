package controllers

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	modeljobsv1alpha1 "github.com/kleveross/klever-model-registry/pkg/apis/modeljob/v1alpha1"
	"github.com/kleveross/klever-model-registry/pkg/common"
)

func getFrameworkByFormat(format modeljobsv1alpha1.Format) modeljobsv1alpha1.Framework {
	return ModelFormatToFrameworkMapping[format]
}

// getORMBDomain is get domain for ModelJob task.
// For model extraction, when it is completed, it should push model to harbor directly
// so that not create ModelJob repeatedly, but for model conversion, when convert complete,
// it should push to klever-model-registry, so that it can extract model automatically.
func getORMBDomain(isConvert bool) string {
	ormbDomain := viper.GetString(common.ORMBDomainEnvKey)
	if isConvert {
		ormbDomain = viper.GetString(KleverModelRegistryAddressEnvKey)
	}

	return ormbDomain
}

// replaceModelRefDomain will replace modelRef domain.
// The real domain is depend on getORMBDomain.
func replaceModelRefDomain(inputModelRef, ormbDomain string) (string, error) {
	refSlice := strings.Split(inputModelRef, "/")

	modelRef := ""
	if len(refSlice) == 2 {
		// The form like release/savedmodel:v1.0, do not have default domain.
		modelRef = strings.Join([]string{ormbDomain, refSlice[0], refSlice[1]}, "/")
	} else if len(refSlice) == 3 {
		// The form like harbor.io/release/savedmodel:v1.0, it have default domain.
		modelRef = strings.Join([]string{ormbDomain, refSlice[1], refSlice[2]}, "/")
	} else {
		return "", fmt.Errorf("The model ref's format is error")
	}

	return modelRef, nil
}

func generateJobResource(modeljob *modeljobsv1alpha1.ModelJob) (*batchv1.Job, error) {
	var dstFormat modeljobsv1alpha1.Format
	var dstFramework modeljobsv1alpha1.Framework
	var srcFormat modeljobsv1alpha1.Format
	var image string
	var srcModelRef string
	var dstModelRef string
	var ormbDomain string
	var err error

	if modeljob.Spec.Conversion != nil {
		if modeljob.Spec.DesiredTag == nil {
			return nil, fmt.Errorf("modeljob desired tag is nil")
		}
		ormbDomain = getORMBDomain(true)
		dstModelRef, err = replaceModelRefDomain(*modeljob.Spec.DesiredTag, ormbDomain)
		if err != nil {
			return nil, err
		}
		dstFormat = modeljob.Spec.Conversion.MMdnn.To
		dstFramework = getFrameworkByFormat(dstFormat)
		srcFormat = modeljob.Spec.Conversion.MMdnn.From
		if imageEnv, ok := presetImage[strings.ToLower(string(modeljob.Spec.Conversion.MMdnn.From))+"-convert"]; ok {
			image = viper.GetString(imageEnv)
		}
		if image == "" {
			return nil, fmt.Errorf("failed get %v model convert image", modeljob.Spec.Conversion.MMdnn.From)
		}
	} else if modeljob.Spec.Extraction != nil {
		ormbDomain = getORMBDomain(false)
		dstModelRef = "empty"
		dstFormat = modeljob.Spec.Extraction.Format
		dstFramework = getFrameworkByFormat(dstFormat)
		srcFormat = dstFormat
		if imageEnv, ok := presetImage[strings.ToLower(string(dstFormat))+"-extract"]; ok {
			image = viper.GetString(imageEnv)
		}
		if image == "" {
			return nil, fmt.Errorf("failed get %v model extract image", dstFormat)
		}
	} else {
		return nil, fmt.Errorf("%v", "not support source")
	}

	srcModelRef, err = replaceModelRefDomain(modeljob.Spec.Model, ormbDomain)
	if err != nil {
		return nil, err
	}

	initContainers, err := generateInitContainers(modeljob)
	if err != nil {
		return nil, err
	}

	cpu, mem := "", ""
	for _, val := range modeljob.Spec.Env {
		if val.Name == ModelJobTaskCPUEnvKey {
			cpu = val.Value
		}
		if val.Name == ModelJobTaskMEMEnvKey {
			mem = val.Value
		}
	}
	resources := generateResources(cpu, mem)

	schedulerName := getSchedulerName()
	backoffLimit := int32(0)
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: modeljob.Namespace,
			Name:      modeljob.Name,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					InitContainers: initContainers,
					Containers: []corev1.Container{
						{
							Name:            "executor",
							Image:           image,
							WorkingDir:      ModelJobWorkDir,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								corev1.EnvVar{
									Name:  modeljobsv1alpha1.FrameworkEnvKey,
									Value: string(dstFramework),
								},
								corev1.EnvVar{
									Name:  modeljobsv1alpha1.FormatEnvKey,
									Value: string(dstFormat),
								},
								corev1.EnvVar{
									Name:  modeljobsv1alpha1.SourceFormatEnvKey,
									Value: string(srcFormat),
								},
								corev1.EnvVar{
									Name:  modeljobsv1alpha1.SourceModelTagEnvKey,
									Value: srcModelRef,
								},
								corev1.EnvVar{
									Name:  modeljobsv1alpha1.DestinationModelTagEnvKey,
									Value: dstModelRef,
								},
								corev1.EnvVar{
									Name:  modeljobsv1alpha1.SourceModelPathEnvKey,
									Value: modeljobsv1alpha1.SourceModelPath,
								},
								corev1.EnvVar{
									Name:  modeljobsv1alpha1.DestinationModelPathEnvKey,
									Value: modeljobsv1alpha1.DestinationModelPath,
								},
								corev1.EnvVar{
									Name:  modeljobsv1alpha1.ExtractorEnvKey,
									Value: strings.ToLower(string(dstFormat)),
								},
								corev1.EnvVar{
									Name:  common.ORMBDomainEnvKey,
									Value: ormbDomain,
								},
								corev1.EnvVar{
									Name:  common.ORMBUsernameEnvkey,
									Value: viper.GetString(common.ORMBUsernameEnvkey),
								},
								corev1.EnvVar{
									Name:  common.ORMBPasswordEnvKey,
									Value: viper.GetString(common.ORMBPasswordEnvKey),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      ModelJobSharedVolumeName,
									MountPath: modeljobsv1alpha1.SourceModelPath,
								},
							},
							Resources: *resources,
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: ModelJobSharedVolumeName,
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
					SchedulerName: schedulerName,
				},
			},
			BackoffLimit: &backoffLimit,
		},
	}

	job.Spec.Template.Spec.Containers[0].Env = append(job.Spec.Template.Spec.Containers[0].Env, modeljob.Spec.Env...)

	return job, nil
}

// generateInitContainers will pull model from harbor and export the model to /models/input path
func generateInitContainers(modeljob *modeljobsv1alpha1.ModelJob) ([]corev1.Container, error) {
	if modeljob.Spec.InitContainer != nil {
		return modeljob.Spec.InitContainer, nil
	}

	ormbDomain := viper.GetString(common.ORMBDomainEnvKey)
	ormbUsername := viper.GetString(common.ORMBUsernameEnvkey)
	ormbPassword := viper.GetString(common.ORMBPasswordEnvKey)
	if ormbDomain == "" || ormbUsername == "" || ormbPassword == "" {
		return nil, nil
	}

	var image string
	if imageEnv, ok := presetImage["initializer"]; ok {
		image = viper.GetString(imageEnv)
	}
	if image == "" {
		return nil, fmt.Errorf("failed get ormb-storage-initializer image")
	}

	cpu, mem := "", ""
	for _, val := range modeljob.Spec.Env {
		if val.Name == ModelInitializerCPUEnvKey {
			cpu = val.Value
		}
		if val.Name == ModelInitializerMEMEnvKey {
			mem = val.Value
		}
	}
	resources := generateResources(cpu, mem)

	initContainers := []corev1.Container{
		{
			Name:  "model-initializer",
			Image: image,
			// Set --relayout=false, only pull and export model, not move any file
			// please refenrence https://github.com/kleveross/ormb/blob/master/cmd/ormb-storage-initializer/cmd/pull-and-export.go
			Args:       []string{modeljob.Spec.Model, modeljobsv1alpha1.SourceModelPath, "--relayout=false"},
			WorkingDir: ModelJobWorkDir,
			Env: []corev1.EnvVar{
				corev1.EnvVar{
					Name:  "ORMB_USERNAME",
					Value: ormbUsername,
				},
				corev1.EnvVar{
					Name:  "ORMB_PASSWORD",
					Value: ormbPassword,
				},
			},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      ModelJobSharedVolumeName,
					MountPath: modeljobsv1alpha1.SourceModelPath,
				},
			},
			Resources:       *resources,
			ImagePullPolicy: corev1.PullAlways,
		},
	}

	return initContainers, nil
}

func getSchedulerName() string {
	schedulerName := viper.GetString(SchedulerNameEnvKey)
	if schedulerName == "" {
		schedulerName = DefaultSchedulerName
	}

	return schedulerName
}

func generateResources(cpu, mem string) *corev1.ResourceRequirements {
	if cpu == "" || mem == "" {
		return &corev1.ResourceRequirements{}
	}

	cpuQuantity := resource.MustParse(cpu)
	memQuantity := resource.MustParse(mem)

	resourceList := corev1.ResourceList{
		corev1.ResourceCPU:    cpuQuantity,
		corev1.ResourceMemory: memQuantity,
	}
	return &corev1.ResourceRequirements{
		Limits:   resourceList,
		Requests: resourceList,
	}
}
