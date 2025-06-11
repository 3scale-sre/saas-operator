package util

import (
	"context"
	"fmt"

	operatorutils "github.com/3scale-sre/saas-operator/internal/pkg/util"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/utils/ptr"
)

const (
	minioPort uint32 = 9000
	region    string = "us-east-1"
)

func MinioClient(ctx context.Context, cfg *rest.Config, podKey types.NamespacedName, user, passwd string) (*s3.Client, chan struct{}, error) {
	localPort, stopCh, err := PortForward(cfg, podKey, minioPort)
	if err != nil {
		return nil, nil, err
	}

	client, err := operatorutils.S3Client(ctx, user, passwd, region, ptr.To(fmt.Sprintf("http://localhost:%d", localPort)))
	if err != nil {
		return nil, nil, err
	}

	return client, stopCh, nil
}
