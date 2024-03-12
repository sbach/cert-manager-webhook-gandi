package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"

	extapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd"

	"github.com/go-gandi/go-gandi"
	"github.com/go-gandi/go-gandi/config"
)

var GroupName = os.Getenv("GROUP_NAME")

func main() {
	if GroupName == "" {
		panic("GROUP_NAME must be specified")
	}

	cmd.RunWebhookServer(GroupName, &GandiDNSProviderSolver{})
}

type GandiDNSProviderSolver struct {
	client *kubernetes.Clientset
}

type SecretSelector struct {
	Name string `json:"name"`
	Key  string `json:"key,omitempty"`
}

type GandiDNSProviderConfig struct {
	APIKeySecretReference              *SecretSelector `json:"apiKeySecretReference"`
	PersonalAccessTokenSecretReference *SecretSelector `json:"personalAccessTokenSecretReference"`
}

func (c *GandiDNSProviderSolver) Name() string {
	return "gandi"
}

func (c *GandiDNSProviderSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	klog.V(6).Infof("Present ACME challenge solving parameters: namespace=%s, zone=%s, fqdn=%s", ch.ResourceNamespace, ch.ResolvedZone, ch.ResolvedFQDN)

	config, err := clientConfig(c.client, ch)
	if err != nil {
		return err
	}

	client := gandi.NewLiveDNSClient(config)

	domain := strings.TrimSuffix(ch.ResolvedZone, ".")
	entry := strings.TrimSuffix(ch.ResolvedFQDN, ch.ResolvedZone)
	entry = strings.TrimSuffix(entry, ".")
	value := []string{ch.Key}

	_, err = client.UpdateDomainRecordByNameAndType(domain, entry, "TXT", 300, value)
	if err != nil {
		return err
	}

	return nil
}

func (c *GandiDNSProviderSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	klog.V(6).Infof("Clean-up ACME challenge record: namespace=%s, zone=%s, fqdn=%s", ch.ResourceNamespace, ch.ResolvedZone, ch.ResolvedFQDN)

	config, err := clientConfig(c.client, ch)
	if err != nil {
		return err
	}

	client := gandi.NewLiveDNSClient(config)

	domain := strings.TrimSuffix(ch.ResolvedZone, ".")
	entry := strings.TrimSuffix(ch.ResolvedFQDN, ch.ResolvedZone)
	entry = strings.TrimSuffix(entry, ".")

	_, err = client.GetDomainRecordByNameAndType(domain, entry, "TXT")
	if err != nil {
		return err
	}

	err = client.DeleteDomainRecord(domain, entry, "TXT")
	if err != nil {
		return err
	}

	return nil
}

func (c *GandiDNSProviderSolver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	cl, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return err
	}

	c.client = cl

	return nil
}

func webhookConfig(cfgJSON *extapi.JSON) (GandiDNSProviderConfig, error) {
	config := GandiDNSProviderConfig{}
	if cfgJSON == nil {
		return config, nil
	}
	if err := json.Unmarshal(cfgJSON.Raw, &config); err != nil {
		return config, fmt.Errorf("error decoding solver configuration: %v", err)
	}

	return config, nil
}

func getSecret(client *kubernetes.Clientset, namespace string, secretName string, keyName string) ([]byte, error) {

	secret, err := client.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get secret `%s` in namespace `%s`: %v", secretName, namespace, err)
	}

	keyBytes, found := secret.Data[keyName]
	if !found {
		return nil, fmt.Errorf("key `%q` not found in secret data", keyName)
	}

	return keyBytes, nil
}

func clientConfig(client *kubernetes.Clientset, ch *v1alpha1.ChallengeRequest) (config.Config, error) {
	clientConfig := config.Config{}
	clientConfig.Debug = true

	config, err := webhookConfig(ch.Config)
	if err != nil {
		return clientConfig, err
	}

	/* Configuration provided a Personal Access Token to use. */
	if config.PersonalAccessTokenSecretReference != nil {

		klog.V(6).Infof("Trying to load Gandi Personal Access Token from secret `%s` in namespace `%s`",
			config.PersonalAccessTokenSecretReference.Name,
			ch.ResourceNamespace)

		value, err := getSecret(client, ch.ResourceNamespace,
			config.PersonalAccessTokenSecretReference.Name,
			config.PersonalAccessTokenSecretReference.Key)
		if err != nil {
			return clientConfig, err
		}

		clientConfig.PersonalAccessToken = string(value)

	/* Configuration provided an API key to use. */
	} else if config.APIKeySecretReference != nil {

		klog.V(6).Infof("Trying to load Gandi API key from secret `%s` in namespace `%s`",
			config.APIKeySecretReference.Name,
			ch.ResourceNamespace)

		value, err := getSecret(client, ch.ResourceNamespace,
			config.APIKeySecretReference.Name,
			config.APIKeySecretReference.Key)
		if err != nil {
			return clientConfig, err
		}

		clientConfig.APIKey = string(value)

	/* No API key or Personal Access Token provided. */
	} else {
		return clientConfig, fmt.Errorf("no API key nor Personal Access Token provided")
	}

	return clientConfig, nil
}
