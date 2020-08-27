package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/zclconf/go-cty/cty"
)

func (a *Aws) getElasticacheClusterImportName(name string, resourceParameters map[string]cty.Value) (string, error) {
	possibleClusterID := resourceParameters["cluster_id"]
	if possibleClusterID.Type() != cty.String {
		return "", fmt.Errorf("skipping %v because its cluster id was not a string but %v", name, possibleClusterID.Type())
	}
	clusterID := possibleClusterID.AsString()

	svc := elasticache.New(a.session)
	input := elasticache.DescribeCacheClustersInput{CacheClusterId: &clusterID}
	output, err := svc.DescribeCacheClusters(&input)
	if err != nil {
		return "", err
	}
	for _, cluster := range output.CacheClusters {
		if *cluster.CacheClusterId == clusterID {
			return clusterID, nil
		}
	}
	return "", fmt.Errorf("aws_elasticache_cluster not found %v", clusterID)

}

func (a *Aws) getServiceDiscoveryServiceImportName(name string) (string, error) {
	svc := servicediscovery.New(a.session)
	input := servicediscovery.ListServicesInput{}
	output, err := svc.ListServices(&input)
	if err != nil {
		return "", err
	}
	for _, service := range output.Services {
		if *service.Name == name {
			return *service.Id, nil
		}
	}
	return "", fmt.Errorf("aws_service_discovery_service not found %v", name)
}

func (a *Aws) getIamRoleImportName(name string) (string, error) {
	svc := iam.New(a.session)

	input := iam.GetRoleInput{RoleName: &name}
	_, err := svc.GetRole(&input)
	if err != nil {
		return "", nil
	}

	return name, nil
}

func (a *Aws) getAppAutoScalingPolicyImportName(name string, resourceParameters map[string]cty.Value) (string, error) {
	svc := applicationautoscaling.New(a.session)

	serviceNamespace := resourceParameters["identifier"]
	if serviceNamespace.Type() != cty.String {
		return "", fmt.Errorf("skipping %v because its service namespace was not a string but %v", name, serviceNamespace.Type())
	}

	namespace := serviceNamespace.AsString()

	input := applicationautoscaling.DescribeScalingPoliciesInput{PolicyNames: []*string{&name}, ServiceNamespace: &namespace}
	output, err := svc.DescribeScalingPolicies(&input)
	if err != nil {
		return "", err
	}

	for _, policy := range output.ScalingPolicies {
		if *policy.PolicyName == name {
			return name, nil
		}
	}
	return "", fmt.Errorf("aws_appautoscaling_policy not found %v", name)
}

func (a *Aws) getRdsInstance(name string, resourceParameters map[string]cty.Value) (string, error) {
	svc := rds.New(a.session)
	possibleIdentifier := resourceParameters["identifier"]
	if possibleIdentifier.Type() != cty.String {
		return "", fmt.Errorf("skipping %v because its identifier was not a string but %v", name, possibleIdentifier.Type().FriendlyName())
	}

	identifier := possibleIdentifier.AsString()
	input := rds.DescribeDBInstancesInput{DBInstanceIdentifier: &identifier}
	output, err := svc.DescribeDBInstances(&input)
	if err != nil {
		return "", err
	}
	for _, instance := range output.DBInstances {
		if *instance.DBInstanceIdentifier == identifier {
			return identifier, nil
		}
	}
	return "", fmt.Errorf("aws_db_instance not found %v", name)
}

func (a *Aws) getEcsServiceImportName(name string, resourceParameters map[string]cty.Value) (string, error) {
	svc := ecs.New(a.session)
	possibleCluster := resourceParameters["cluster"]
	if possibleCluster.Type() != cty.String {
		return "", fmt.Errorf("skipping %v because its cluster was not a string but %v", name, possibleCluster.Type().FriendlyName())
	}
	cluster := possibleCluster.AsString()
	input := ecs.DescribeServicesInput{Cluster: &cluster, Services: []*string{&name}}
	output, err := svc.DescribeServices(&input)
	if err != nil {
		return "", nil
	}
	for _, service := range output.Services {
		clusterName := strings.Split(cluster, "/")
		if *service.ServiceName == name {
			return fmt.Sprintf("%s/%s", clusterName[len(clusterName)-1], name), nil
		}
	}
	return "", fmt.Errorf("aws_ecs_service not found %v", name)
}

func (a *Aws) getSecurityGroupImportName(name string, resourceParameters map[string]cty.Value) (string, error) {
	svc := ec2.New(a.session)
	possibleVpcID := resourceParameters["vpc_id"]
	if possibleVpcID.Type() != cty.String {
		return "", fmt.Errorf("skipping %v because its vpc id was not a string but %v", name, possibleVpcID.Type().FriendlyName())
	}
	vpcID := possibleVpcID.AsString()

	vpcKey := "vpc-id"
	vpcFilter := ec2.Filter{Name: &vpcKey, Values: []*string{&vpcID}}

	nameKey := "group-name"
	nameFilter := ec2.Filter{Name: &nameKey, Values: []*string{&name}}

	input := ec2.DescribeSecurityGroupsInput{Filters: []*ec2.Filter{&vpcFilter, &nameFilter}}
	output, err := svc.DescribeSecurityGroups(&input)

	if err != nil {
		return "", err
	}
	for _, securityGroup := range output.SecurityGroups {
		if *securityGroup.GroupName == name {
			return *securityGroup.GroupId, nil
		}
	}
	return "", fmt.Errorf("aws_security_group not found %v", name)
}

func (a *Aws) getSsmParameterImportName(name string) (string, error) {
	svc := ssm.New(a.session)
	input := ssm.GetParameterInput{Name: &name}
	output, err := svc.GetParameter(&input)
	if err != nil {
		return "", nil
	}
	return *output.Parameter.Name, nil
}

func (a *Aws) getEcrImportName(name string) (string, error) {
	svc := ecr.New(a.session)
	input := ecr.DescribeRepositoriesInput{RepositoryNames: []*string{&name}}
	output, err := svc.DescribeRepositories(&input)
	if err != nil {
		return "", nil
	}
	for _, repository := range output.Repositories {
		if *repository.RepositoryName == name {
			return name, nil
		}
	}
	return "", fmt.Errorf("aws_ecr_repository not found %v", name)
}

func (a *Aws) getCloudWatchGroupImportName(name string) (string, error) {
	svc := cloudwatchlogs.New(a.session)
	input := cloudwatchlogs.DescribeLogGroupsInput{LogGroupNamePrefix: &name}
	output, err := svc.DescribeLogGroups(&input)
	if err != nil {
		return "", nil
	}
	for _, group := range output.LogGroups {
		if *group.LogGroupName == name {
			return name, nil
		}
	}
	return "", fmt.Errorf("aws_cloudwatch_log_group not found %v", name)
}
