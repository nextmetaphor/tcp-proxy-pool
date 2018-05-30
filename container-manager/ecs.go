package container_manager

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"time"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/nextmetaphor/tcp-proxy-pool/log"
	"github.com/sirupsen/logrus"
	"github.com/nextmetaphor/tcp-proxy-pool/application"
	"github.com/nextmetaphor/tcp-proxy-pool/container"
)

const (
	maximumContainerStartTimeSecDefault       = 60
	logAWSErrorOccurred                       = "AWS error occurred"
	logNonAWSErrorOccurred                    = "Non-AWS error occurred"
	logRunTaskOutput                          = "RunTask output"
	logWaitingForTaskNetworkInterfaceToAttach = "Waiting for task [%s] network interface to attach, timing out in [%n] second(s)"
	logDescribeTaskOutput                     = "DescribeTask output"
	logTaskNetworkInterfaceStatus             = "Task [%s] network interface in state [%s]"
)

type (
	ECS struct {
		logger logrus.Logger
		conf application.ECSSettings
	}
)

func strArrToStrPointerArr(strArr []string) []*string {
	ps := make([]*string, len(strArr))
	for i, s := range strArr {
		*ps[i] = s
	}
	return ps
}

func (cm ECS) CreateContainer() (*container.Container, error) {
	config := &aws.Config{Region: aws.String(cm.conf.Region)}
	if cm.conf.Profile != "" {
		config.Credentials = credentials.NewSharedCredentials("", cm.conf.Profile)
	}
	session, err := session.NewSession(config)

	if err != nil {
		return nil, err
	}

	ecsService := ecs.New(session)
	runTaskInput := &ecs.RunTaskInput{
		Cluster:        aws.String(cm.conf.Cluster),
		TaskDefinition: aws.String(cm.conf.TaskDefinition),
		Count:          aws.Int64(1),
		LaunchType:     aws.String(cm.conf.LaunchType),
		NetworkConfiguration: &ecs.NetworkConfiguration{
			AwsvpcConfiguration: &ecs.AwsVpcConfiguration{
				AssignPublicIp: aws.String(cm.conf.AssignPublicIP),
				SecurityGroups: strArrToStrPointerArr(cm.conf.SecurityGroups),
				Subnets:        strArrToStrPointerArr(cm.conf.Subnets),
			},
		},
	}

	runTaskOutput, err := ecsService.RunTask(runTaskInput)
	if err != nil {
		if err, ok := err.(awserr.Error); ok {
			switch err.Code() {
			case ecs.ErrCodeServerException:
				log.LogError(ecs.ErrCodeServerException, err, &cm.logger)
			case ecs.ErrCodeClientException:
				log.LogError(ecs.ErrCodeClientException, err, &cm.logger)
			case ecs.ErrCodeInvalidParameterException:
				log.LogError(ecs.ErrCodeInvalidParameterException, err, &cm.logger)
			case ecs.ErrCodeClusterNotFoundException:
				log.LogError(ecs.ErrCodeClusterNotFoundException, err, &cm.logger)
			default:
				log.LogError(logAWSErrorOccurred, err, &cm.logger)
			}
		} else {
			log.LogError(logNonAWSErrorOccurred, err, &cm.logger)
		}
		return nil, err
	}

	cm.logger.Info(logWaitingForTaskNetworkInterfaceToAttach, cm.conf.MaximumContainerStartTimeSec, *runTaskOutput.Tasks[0].TaskArn)
	cm.logger.Debug(logRunTaskOutput, runTaskOutput)

	maximumStartTimeSec := cm.conf.MaximumContainerStartTimeSec
	if maximumStartTimeSec <= 0 {
		maximumStartTimeSec = maximumContainerStartTimeSecDefault
	}

	// TODO horribly brittle - need to check array sizes here
	var describeTasksOutput *ecs.DescribeTasksOutput
	for i := 0; i < maximumStartTimeSec; i++ {
		time.Sleep(1 * time.Second)
		describeTasksOutput, err = ecsService.DescribeTasks(&ecs.DescribeTasksInput{
			Tasks:   []*string{runTaskOutput.Tasks[0].TaskArn},
			Cluster: runTaskInput.Cluster,
		})
		if err != nil {
			return nil, err
		}

		cm.logger.Info(logTaskNetworkInterfaceStatus, *runTaskOutput.Tasks[0].TaskArn, *describeTasksOutput.Tasks[0].Attachments[0].Status)
		cm.logger.Debug(logDescribeTaskOutput, describeTasksOutput.Tasks[0])

		fmt.Println(*describeTasksOutput.Tasks[0].Attachments[0].Status)
		if *describeTasksOutput.Tasks[0].Attachments[0].Status == "ATTACHED" {
			break
		}
	}

	return &container.Container{
		ExternalID: *runTaskOutput.Tasks[0].TaskArn,
		StartTime:  *runTaskOutput.Tasks[0].CreatedAt,
		IPAddress:  *describeTasksOutput.Tasks[0].Containers[0].NetworkInterfaces[0].PrivateIpv4Address,
		Port:       8080,
	}, nil
}

func (cm ECS) DestroyContainer(externalID string) (error) {
	// TODO obvs needs implementing
	return nil
}
