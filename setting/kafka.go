package setting

type KafkaSettings struct {
	Topics    []string
	Consumer  KafkaConsumerSettings
	Zookeeper KafkaZookeeperSettings
}

type KafkaConsumerSettings struct {
	Group  string
	Buffer int
	Offset string
}

type KafkaZookeeperSettings struct {
	Servers []string
	Root    string
}
