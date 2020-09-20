package aws

import (
	"fmt"
	"os"

	"github.com/jaysonsantos/terraform-plan-importer/importer"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/zclconf/go-cty/cty"
)

type Aws struct {
	session *session.Session
}

func init() {
	importer.RegisterImporter(&Aws{})
}

func New(region *string) *Aws {
	sess := session.Must(session.NewSession(&aws.Config{Region: region}))
	return &Aws{session: sess}
}

func (a *Aws) Init() error {
	region, ok := os.LookupEnv("AWS_REGION")
	if !ok {
		region = "eu-central-1"
	}
	a.session = session.Must(session.NewSession(&aws.Config{Region: &region}))
	return nil
}

func (a *Aws) GetImportName(resourceType string, name string, resourceParameters map[string]cty.Value) (string, error) {
	switch resourceType {
	case "aws_elasticache_cluster":
		return a.getElasticacheClusterImportName(name, resourceParameters)
	case "aws_cloudwatch_log_group":
		return a.getCloudWatchGroupImportName(name)
	case "aws_ecr_repository":
		return a.getEcrImportName(name)
	case "aws_ssm_parameter":
		return a.getSsmParameterImportName(name)
	case "aws_security_group":
		return a.getSecurityGroupImportName(name, resourceParameters)
	case "aws_ecs_service":
		return a.getEcsServiceImportName(name, resourceParameters)
	case "aws_db_instance":
		return a.getRdsInstance(name, resourceParameters)
	case "aws_appautoscaling_policy":
		return a.getAppAutoScalingPolicyImportName(name, resourceParameters)
	case "aws_iam_role":
		return a.getIamRoleImportName(name)
	case "aws_service_discovery_service":
		return a.getServiceDiscoveryServiceImportName(name)
	}

	return "", fmt.Errorf("resourceType not supported %v", resourceType)
}

func (a *Aws) ImporterName() string {
	return "aws"
}
