package functests

import (
	"github.com/onsi/gomega"

	"fmt"
	"time"

	k8sbatchv1 "k8s.io/api/batch/v1"
	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	utilwait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

// WaitForPVCBound waits for a pvc with a given name and namespace to reach BOUND phase
func WaitForPVCBound(k8sClient *kubernetes.Clientset, pvcName string, pvcNamespace string) {
	gomega.Eventually(func() error {
		pvc, err := k8sClient.CoreV1().PersistentVolumeClaims(pvcNamespace).Get(pvcName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if pvc.Status.Phase == k8sv1.ClaimBound {
			return nil
		}
		return fmt.Errorf("Waiting on pvc %s/%s to reach bound state when it is currently %s", pvcNamespace, pvcName, pvc.Status.Phase)
	}, 200*time.Second, 1*time.Second).ShouldNot(gomega.HaveOccurred())
}

// WaitForJobSucceeded waits for a Job with a given name and namespace to succeed until 200 seconds
func WaitForJobSucceeded(k8sClient *kubernetes.Clientset, jobName string, jobNamespace string) {
	gomega.Eventually(func() error {
		job, err := k8sClient.BatchV1().Jobs(jobNamespace).Get(jobName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if job.Status.Succeeded > 0 {
			return nil
		}
		return fmt.Errorf("Waiting on job %s/%s to succeed when it is currently %d", jobName, jobNamespace, job.Status.Succeeded)
	},
		200*time.Second, 1*time.Second).Should(gomega.Succeed())
}

// GetRandomPVC returns a pvc with a randomized name
func GetRandomPVC(storageClass string, quantity string) *k8sv1.PersistentVolumeClaim {
	storageQuantity, err := resource.ParseQuantity(quantity)
	gomega.Expect(err).To(gomega.BeNil())

	randomName := "test-pvc-" + rand.String(12)

	pvc := &k8sv1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolumeClaim",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: randomName,
		},
		Spec: k8sv1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClass,
			AccessModes:      []k8sv1.PersistentVolumeAccessMode{k8sv1.ReadWriteOnce},

			Resources: k8sv1.ResourceRequirements{
				Requests: k8sv1.ResourceList{
					"storage": storageQuantity,
				},
			},
		},
	}

	return pvc
}

// GetDataValidatorJob returns the spec of a job
func GetDataValidatorJob(pvc string) *k8sbatchv1.Job {
	randomName := "test-job-" + rand.String(12)
	job := &k8sbatchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: randomName,
		},
		Spec: k8sbatchv1.JobSpec{
			Template: k8sv1.PodTemplateSpec{
				Spec: k8sv1.PodSpec{
					RestartPolicy: k8sv1.RestartPolicyNever,
					Containers: []k8sv1.Container{
						k8sv1.Container{
							Name:  randomName,
							Image: "busybox",
							VolumeMounts: []k8sv1.VolumeMount{
								k8sv1.VolumeMount{
									MountPath: "/data",
									Name:      "volume-to-debug",
								},
							},
							Command: []string{"/bin/sh", "-c"},
							Args: []string{
								"dd if=/dev/zero of=/tmp/random.img bs=512 count=1",       //This command creates new file named random.img
								"md5VAR1=$(md5sum /tmp/random.img | awk '{ print $1 }')",  //calculates md5sum of random.img and stores it in a variable
								"cp /tmp/random.img /data/random.img",                     //copies random.img file to pvc's mountpoint
								"md5VAR2=$(md5sum /data/random.img | awk '{ print $1 }')", //calculates md5sum of file random.img
								"if [[ \"$md5VAR1\" != \"$md5VAR2\" ]];then exit 1; fi",   //compares the md5sum of random.img file with previous one
							},
						},
					},
					Volumes: []k8sv1.Volume{
						k8sv1.Volume{
							Name: "volume-to-debug",
							VolumeSource: k8sv1.VolumeSource{
								PersistentVolumeClaim: &k8sv1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvc,
								},
							},
						},
					},
				},
			},
		},
	}
	return job
}

// DeleteNamespaceAndWait deletes a namespace and waits on it to terminate
func (t *DeployManager) DeleteNamespaceAndWait(namespace string) error {
	err := t.DeleteStorageClusterAndWait(namespace)
	if err != nil {
		return err
	}
	err = t.k8sClient.CoreV1().Namespaces().Delete(namespace, &metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	timeout := 600 * time.Second
	interval := 10 * time.Second

	// Wait for namespace to terminate
	err = utilwait.PollImmediate(interval, timeout, func() (done bool, err error) {
		_, err = t.k8sClient.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
		if !errors.IsNotFound(err) {
			return false, nil
		}
		return true, nil
	})

	return err
}

// GetDeploymentImage returns the deployment image name for the deployment
func (t *DeployManager) GetDeploymentImage(name string) (string, error) {
	deployment, err := t.k8sClient.AppsV1().Deployments(InstallNamespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return deployment.Spec.Template.Spec.Containers[0].Image, nil
}

// CreateNamespace creates a namespace in the cluster, ignoring if it already exists
func (t *DeployManager) CreateNamespace(namespace string) error {
	label := make(map[string]string)
	// Label required for monitoring this namespace
	label["openshift.io/cluster-monitoring"] = "true"
	ns := &k8sv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   namespace,
			Labels: label,
		},
	}
	_, err := t.k8sClient.CoreV1().Namespaces().Create(ns)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

// DeleteStorageClusterAndWait deletes a storageClusterCR and waits on it to terminate
func (t *DeployManager) DeleteStorageClusterAndWait(namespace string) error {
	err := t.deleteStorageCluster()
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	cephClusters, err := t.rookClient.CephV1().CephClusters(namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, cephCluster := range cephClusters.Items {
		_, err = t.rookClient.CephV1().CephClusters(namespace).Patch(cephCluster.GetName(), types.JSONPatchType, []byte(finalizerRemovalPatch))
		if err != nil {
			return err
		}
	}

	timeout := 600 * time.Second
	interval := 10 * time.Second

	// Wait for storagecluster and cephCluster to terminate
	err = utilwait.PollImmediate(interval, timeout, func() (done bool, err error) {
		cephClusters, err := t.rookClient.CephV1().CephClusters(namespace).List(metav1.ListOptions{})
		if err != nil {
			return false, err
		}
		if len(cephClusters.Items) != 0 {
			return false, nil
		}
		_, err = t.getStorageCluster()
		if !errors.IsNotFound(err) {
			return false, nil
		}
		return true, nil
	})

	return err
}
