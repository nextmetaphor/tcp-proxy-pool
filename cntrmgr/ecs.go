package cntrmgr

import (
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"time"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/nextmetaphor/tcp-proxy-pool/log"
	"github.com/sirupsen/logrus"
	"github.com/nextmetaphor/tcp-proxy-pool/application"
	"github.com/nextmetaphor/tcp-proxy-pool/cntr"
)

const (
	maximumContainerStartTimeSecDefault       = 60
	logAWSErrorOccurred                       = "AWS error occurred"
	logNonAWSErrorOccurred                    = "Non-AWS error occurred"
	logRunTaskOutput                          = "RunTask output"
	logWaitingForTaskNetworkInterfaceToAttach = "Waiting for task [%s] network interface to attach, timing out in [%d] second(s)"
	logDescribeTaskOutput                     = "DescribeTask output"
	logTaskNetworkInterfaceStatus             = "Task [%s] network interface in state [%s]"
)

type (
	ECS struct {
		Logger     logrus.Logger
		Conf       application.ECSSettings
		ECSService *ecs.ECS
	}
)

func strArrToStrPointerArr(strArr []string) []*string {
	ps := make([]*string, len(strArr))
	for i:= 0; i<len(strArr); i++ {
		ps[i] = &strArr[i]
	}
	return ps
}

func (cm *ECS) InitialiseECSService() (error) {
	config := &aws.Config{Region: aws.String(cm.Conf.Region)}
	if cm.Conf.Profile != "" {
		config.Credentials = credentials.NewSharedCredentials("", cm.Conf.Profile)
	}
	session, err := session.NewSession(config)
	if err != nil {
		return err
	}

	cm.ECSService = ecs.New(session)
	return nil
}

func (cm ECS) CreateContainer() (*cntr.Container, error) {
	runTaskInput := &ecs.RunTaskInput{
		Cluster:        aws.String(cm.Conf.Cluster),
		TaskDefinition: aws.String(cm.Conf.TaskDefinition),
		Count:          aws.Int64(1),
		LaunchType:     aws.String(cm.Conf.LaunchType),
		NetworkConfiguration: &ecs.NetworkConfiguration{
			AwsvpcConfiguration: &ecs.AwsVpcConfiguration{
				AssignPublicIp: aws.String(cm.Conf.AssignPublicIP),
				SecurityGroups: strArrToStrPointerArr(cm.Conf.SecurityGroups),
				Subnets:        strArrToStrPointerArr(cm.Conf.Subnets),
			},
		},
	}

	runTaskOutput, err := cm.ECSService.RunTask(runTaskInput)
	if err != nil {
		if err, ok := err.(awserr.Error); ok {
			switch err.Code() {
			case ecs.ErrCodeServerException:
				log.LogError(ecs.ErrCodeServerException, err, &cm.Logger)
			case ecs.ErrCodeClientException:
				log.LogError(ecs.ErrCodeClientException, err, &cm.Logger)
			case ecs.ErrCodeInvalidParameterException:
				log.LogError(ecs.ErrCodeInvalidParameterException, err, &cm.Logger)
			case ecs.ErrCodeClusterNotFoundException:
				log.LogError(ecs.ErrCodeClusterNotFoundException, err, &cm.Logger)
			default:
				log.LogError(logAWSErrorOccurred, err, &cm.Logger)
			}
		} else {
			log.LogError(logNonAWSErrorOccurred, err, &cm.Logger)
		}
		return nil, err
	}

	cm.Logger.Infof(logWaitingForTaskNetworkInterfaceToAttach, *runTaskOutput.Tasks[0].TaskArn, cm.Conf.MaximumContainerStartTimeSec)
	cm.Logger.Debug(logRunTaskOutput, runTaskOutput)

	maximumStartTimeSec := cm.Conf.MaximumContainerStartTimeSec
	if maximumStartTimeSec <= 0 {
		maximumStartTimeSec = maximumContainerStartTimeSecDefault
	}

	// TODO horribly brittle - need to check array sizes here
	var describeTasksOutput *ecs.DescribeTasksOutput
	for i := 0; i < maximumStartTimeSec; i++ {
		time.Sleep(1 * time.Second)
		describeTasksOutput, err = cm.ECSService.DescribeTasks(&ecs.DescribeTasksInput{
			Tasks:   []*string{runTaskOutput.Tasks[0].TaskArn},
			Cluster: runTaskInput.Cluster,
		})
		if err != nil {
			return nil, err
		}

		cm.Logger.Infof(logTaskNetworkInterfaceStatus, *runTaskOutput.Tasks[0].TaskArn, *describeTasksOutput.Tasks[0].Attachments[0].Status)
		cm.Logger.Debug(logDescribeTaskOutput, describeTasksOutput.Tasks[0])

		if *describeTasksOutput.Tasks[0].Attachments[0].Status == "ATTACHED" {
			break
		}
	}

	return &cntr.Container{
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