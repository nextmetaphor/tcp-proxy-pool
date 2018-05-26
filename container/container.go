package container

//func LoadAWSConfig() (cfg aws.Config, err error) {
//	cfg, err = external.LoadDefaultAWSConfig()
//	if err == nil {
//		cfg.Region = endpoints.UsWest2RegionID
//	}
//
//	return cfg, err
//}
//
//func GetService(cfg aws.Config, serviceName string) {
//	svc := ecs.New(cfg)
//	input := &ecs.DescribeServicesInput{
//		Services: []string{
//			serviceName,
//		},
//	}
//}
