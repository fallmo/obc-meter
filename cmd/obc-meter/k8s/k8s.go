package k8s

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/fallmo/obc-meter/cmd/obc-meter/db"
	obcv1alpha1 "github.com/kube-object-storage/lib-bucket-provisioner/pkg/apis/objectbucket.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

var client *dynamic.DynamicClient
var scheme = runtime.NewScheme()

var OBCGroupVersionResource = schema.GroupVersionResource{
	Group:    "objectbucket.io",
	Version:  "v1alpha1",
	Resource: "objectbucketclaims",
}

var CMGroupResourceVersion = schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "configmaps",
}

var SCGroupResourceVersion = schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "secrets",
}

func connectToKubernetes() {
	var config *rest.Config
	_ = obcv1alpha1.AddToScheme(scheme)

	if os.Getenv("APP_ENV") == "development" {
		K8S_API_URL := os.Getenv("K8S_API_URL")
		K8S_API_TOKEN := os.Getenv("K8S_API_TOKEN")
		if K8S_API_URL == "" {
			log.Fatalln("Variable 'K8S_API_URL' is required during development")
		}
		if K8S_API_TOKEN == "" {
			log.Fatalln("Variable 'K8S_API_TOKEN' is required during development")
		}
		log.Println("Connecting to Kubernetes with environment variables..")

		config = &rest.Config{
			Host:        K8S_API_URL,
			BearerToken: K8S_API_TOKEN,
		}
	} else {
		log.Println("Connecting to Kubernetes from inside cluster..")
		cfg, err := rest.InClusterConfig()

		if err != nil {
			log.Fatalln(err, "\nFailed to connect to Kubernetes..")
		}

		config = cfg
	}

	var err error
	client, err = dynamic.NewForConfig(config)

	if err != nil {
		fmt.Println(err)
		log.Fatalln("Failed to connect to Kubernetes...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = client.Resource(OBCGroupVersionResource).Namespace("openshift").List(ctx, v1.ListOptions{})
	if err != nil {
		fmt.Println(err)
		log.Fatalln("Failed to list OBCs")
	}

	_, err = client.Resource(CMGroupResourceVersion).Namespace("openshift").List(ctx, v1.ListOptions{})
	if err != nil {
		fmt.Println(err)
		log.Fatalln("Failed to list Configmaps")
	}

	_, err = client.Resource(SCGroupResourceVersion).Namespace("openshift").List(ctx, v1.ListOptions{})
	if err != nil {
		fmt.Println(err)
		log.Fatalln("Failed to list Configmaps")
	}

	log.Printf("Successfully connected to Kubernetes API...")

}

func StartMeteringObjectBuckets() {
	connectToKubernetes()
	meterObjectBuckets("automatic")
}

func meterObjectBuckets(trigger string) {
	log.Println("Running Metering")

	runId, err := db.OpenRun(trigger)
	if err != nil {
		fmt.Println(err)
		log.Fatal("Failed to start a run\n")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	res, err := client.Resource(OBCGroupVersionResource).List(ctx, v1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			getLabelKey(): "true",
		}).String(),
	})

	if err != nil {
		fmt.Println(err)
		log.Fatal("Failed to list object buckets")
	}

	log.Printf("Found '%v' ObjectBucketClaims to meter\n", len(res.Items))

	runSummary := db.CloseRunArgs{
		AllUids:       []string{},
		FailedUids:    []string{},
		ErrorMessages: []string{},
	}

	for i := 0; i < len(res.Items); i++ {
		obc := res.Items[i]
		uid := string(obc.GetUID())

		_, err := meterObjectBucket(obc.GetName(), uid, obc.GetNamespace(), *runId)

		runSummary.AllUids = append(runSummary.AllUids, uid)

		if err != nil {
			runSummary.FailedUids = append(runSummary.FailedUids, uid)
			runSummary.ErrorMessages = append(runSummary.ErrorMessages, err.Error())
		}
	}

	err = db.CloseRun(*runId, runSummary)

	log.Printf("Finished metering '%v' ObjectBucketClaims\n", len(res.Items))
}

func meterObjectBucket(name string, uid string, namespace string, runId int) (bool, error) {
	fmt.Printf("\nMetering Bucket [Name=%v, Uid=%v, Namespace=%v]\n", name, uid, namespace)
	keys, err := getBucketKeys(name, namespace)
	if err != nil {
		fmt.Println(err)
		log.Printf("Failed to meter bucket [Name=%v, Uid=%v, Namespace=%v]\n", name, uid, namespace)
		return false, err
	}

	config, err := getBucketConfig(name, namespace)
	if err != nil {
		fmt.Println(err)
		log.Printf("Failed to meter bucket [Name=%v, Uid=%v, Namespace=%v]\n", name, uid, namespace)
		return false, err
	}

	stats, err := getBucketStats(config, keys)
	if err != nil {
		fmt.Println(err)
		log.Printf("Failed to meter bucket [Name=%v, Uid=%v, Namespace=%v]\n", name, uid, namespace)
		return false, err
	}

	currentRecord, err := db.GetBucketCurrentRecord(uid)
	if err != nil {
		fmt.Println(err)
		log.Printf("Failed to meter bucket [Name=%v, Uid=%v, Namespace=%v]\n", name, uid, namespace)
		return false, err
	}

	// no previous record or previous record is changed
	if currentRecord == nil || stats.bytesTotal != uint(currentRecord.BytesTotal) || stats.objectsCount != uint(currentRecord.ObjectsCount) {
		_, err := db.AppendBucketUsageRecord(db.AppendBucketUsageRecordArgs{
			BucketUid:    uid,
			ObjectsCount: uint64(stats.objectsCount),
			BytesTotal:   uint64(stats.bytesTotal),
			RunId:        runId,
		})

		if err != nil {
			fmt.Println(err)
			log.Printf("Failed to meter bucket [Name=%v, Uid=%v, Namespace=%v]\n", name, uid, namespace)
			return false, err
		}

		log.Printf("Successfully metered bucket (UPDATED) [Name=%v, Uid=%v, Namespace=%v]\n", name, uid, namespace)

		return true, nil

	} else {
		log.Printf("Successfully metered bucket (UNCHANGED) [Name=%v, Uid=%v, Namespace=%v]\n", name, uid, namespace)
		return false, nil
	}

}

type bucketKeys struct {
	accessKeyId string
	secretKey   string
}

func getBucketKeys(secretName string, namespace string) (*bucketKeys, error) {
	ctx := context.TODO()
	obj, err := client.Resource(SCGroupResourceVersion).Namespace(namespace).Get(ctx, secretName, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	secret, err := convertToSecret(obj)
	if err != nil {
		return nil, err
	}

	keys := bucketKeys{
		accessKeyId: string(secret.Data["AWS_ACCESS_KEY_ID"]),
		secretKey:   string(secret.Data["AWS_SECRET_ACCESS_KEY"]),
	}

	if keys.accessKeyId == "" {
		return nil, errors.New("Could not retrieve 'AWS_ACCESS_KEY_ID'")
	}

	if keys.secretKey == "" {
		return nil, errors.New("Could not retrieve 'AWS_SECRET_ACCESS_KEY'")
	}

	return &keys, nil
}

type bucketConfig struct {
	name   string
	host   string
	port   string
	region string
}

func getBucketConfig(configmapName string, namespace string) (*bucketConfig, error) {
	ctx := context.TODO()
	obj, err := client.Resource(CMGroupResourceVersion).Namespace(namespace).Get(ctx, configmapName, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	configmap, err := convertToConfigmap(obj)
	if err != nil {
		return nil, err
	}

	config := bucketConfig{
		name:   configmap.Data["BUCKET_NAME"],
		host:   configmap.Data["BUCKET_HOST"],
		port:   configmap.Data["BUCKET_PORT"],
		region: configmap.Data["BUCKET_REGION"],
	}

	if config.name == "" {
		return nil, errors.New("Could not retrieve 'BUCKET_NAME'")
	}

	if config.host == "" {
		return nil, errors.New("Could not retrieve 'BUCKET_HOST'")
	}
	if config.port == "" {
		return nil, errors.New("Could not retrieve 'BUCKET_PORT'")
	}
	if config.region == "" {
		config.region = "us-east-1"
	}

	return &config, nil
}

type bucketStats struct {
	objectsCount uint
	bytesTotal   uint
}

func getBucketStats(config *bucketConfig, keys *bucketKeys) (*bucketStats, error) {
	sess, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String(getEndpoint(config.host, config.port)),
		Region:           aws.String(config.region),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      credentials.NewStaticCredentials(keys.accessKeyId, keys.secretKey, ""),
	})

	if err != nil {
		return nil, err
	}

	svc := s3.New(sess)

	result, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(config.name)})

	if err != nil {
		return nil, err
	}

	stats := bucketStats{
		objectsCount: 0,
		bytesTotal:   0,
	}

	for i := 0; i < len(result.Contents); i++ {
		obj := result.Contents[i]
		stats.objectsCount += 1
		stats.bytesTotal += uint(*obj.Size)
	}

	return &stats, nil
}

func convertToSecret(obj *unstructured.Unstructured) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, secret)

	return secret, err
}

func convertToConfigmap(obj *unstructured.Unstructured) (*corev1.ConfigMap, error) {
	configmap := &corev1.ConfigMap{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, configmap)

	return configmap, err
}

func getEndpoint(host string, port string) string {
	protocol := "http"
	if strings.Contains(port, "443") {
		protocol += "s"
	}
	return protocol + "://" + host + ":" + port
}

func getLabelKey() string {
	labelKey := os.Getenv("LABEL_KEY")
	if labelKey != "" {
		return labelKey
	} else {
		return "meter-activated"
	}
}
