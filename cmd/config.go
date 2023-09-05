package main

type Config struct {
	Port     int       `yaml:"port"`
	RabbitMQ *RabbitMQ `yaml:"rabbitmq"`
}

type RabbitMQ struct {
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	Exchange   string `yaml:"exchange"`
	RoutingKey string `yaml:"routing_key"`
	Queue      string `yaml:"queue"`
}
