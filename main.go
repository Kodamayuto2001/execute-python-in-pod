package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func CreatePythonPod(kubeconfigPath, scriptPath string) (*v1.Pod, error) {
	// クラスターの設定を取得
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}

	// Kubernetes クライアントを作成
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	absScriptPath, err := filepath.Abs(scriptPath)
	if err != nil {
		return nil, err
	}

	// Pod 定義
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "python-script-",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:    "python",
					Image:   "python:3.8",
					Command: []string{"python", "/scripts/" + filepath.Base(scriptPath)},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "script-volume",
							MountPath: "/scripts",
						},
					},
				},
			},
			Volumes: []v1.Volume{
				{
					Name: "script-volume",
					VolumeSource: v1.VolumeSource{
						HostPath: &v1.HostPathVolumeSource{
							Path: filepath.Dir(absScriptPath),
						},
					},
				},
			},
			RestartPolicy: v1.RestartPolicyNever,
		},
	}

	// Pod を作成
	pod, err = clientset.CoreV1().Pods("default").Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return pod, nil
}

func WaitForPodCompletion(clientset *kubernetes.Clientset, podName string) error {
	for {
		pod, err := clientset.CoreV1().Pods("default").Get(context.TODO(), podName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
}

func GetPodLogs(clientset *kubernetes.Clientset, podName string) (string, error) {
	podLogOpts := v1.PodLogOptions{}
	req := clientset.CoreV1().Pods("default").GetLogs(podName, &podLogOpts)
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		return "", err
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", err
	}
	str := buf.String()

	return str, nil
}

func main() {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	// カレントディレクトリを取得
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// script.py の絶対パスを生成
	absScriptPath := filepath.Join(currentDir, "script.py")
	config, _ := clientcmd.BuildConfigFromFlags("", kubeconfig)
	clientset, _ := kubernetes.NewForConfig(config)

	pod, err := CreatePythonPod(kubeconfig, absScriptPath)
	if err != nil {
		panic(err)
	}

	if err := WaitForPodCompletion(clientset, pod.Name); err != nil {
		panic(err)
	}

	logs, err := GetPodLogs(clientset, pod.Name)
	if err != nil {
		panic(err)
	}

	fmt.Println("Pod logs:\n", logs)
}
