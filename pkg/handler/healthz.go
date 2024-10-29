package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/h4-poc/service/pkg/kube"
)

func Healthz(c *gin.Context) {
	body := map[string]interface{}{
		"status": "ok",
	}

	ok, err := checkKubernetesHealth()
	if !ok {
		body["status"] = "kubernetes health check failed"
		body["error"] = err
	} else {
		body["kubernetes"] = "ok"
	}

	ok, err = checkArogCDHealth()
	if !ok {
		body["status"] = "argocd health check failed"
		body["error"] = err
	} else {
		body["argocd"] = "ok"
	}

	c.JSON(200, body)
}

// no exported function
func checkArogCDHealth() (bool, error) {
	// check the argocd server health

	// list namespaces: argocd
	kubeClient, err := kube.NewClient()
	if err != nil {
		log.Errorf("checkArogCDHealth: %v", err)
		return false, err
	}

	namespaces, err := kubeClient.CoreV1().Namespaces().Get(context.Background(), "argocd", metav1.GetOptions{})
	log.Debugf("checkArogCDHealth: %v", namespaces)
	if err != nil {
		log.Errorf("checkArogCDHealth: %v", err)
		return false, err
	}

	// get the svc
	// !todo
	return true, nil

}

// no exported function
func checkKubernetesHealth() (bool, error) {
	kubeClient, err := kube.NewClient()

	_, err = kubeClient.ServerVersion()
	log.Debugf("checkKubernetesHealth: %v", err)
	if err != nil {
		return false, err
	}

	return true, nil
}
